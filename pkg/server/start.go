package server

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/hashicorp/yamux"
	"github.com/jeffbaumes/gogame/pkg/common"
	_ "github.com/mattn/go-sqlite3" // Needed to use sqlite
)

var (
	universe *common.Universe
)

// Start takes a name, seed, and port and starts the universe server
func Start(name string, seed, port int) {
	if port == 0 {
		port = 5555
	}
	dbName := "worlds/" + name + ".db"

	db, err := sql.Open("sqlite3", dbName)
	checkErr(err)

	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS chunk (planet INT, lon INT, lat INT, alt INT, data BLOB, PRIMARY KEY (planet, lat, lon, alt))")
	checkErr(err)
	_, err = stmt.Exec()
	checkErr(err)
	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS planet (id INT PRIMARY KEY, data BLOB)")
	checkErr(err)
	_, err = stmt.Exec()
	checkErr(err)
	stmt, err = db.Prepare("CREATE TABLE IF NOT EXISTS player (name TEXT PRIMARY KEY, data BLOB)")
	checkErr(err)
	_, err = stmt.Exec()
	checkErr(err)

	universe = common.NewUniverse(db, "sun-moon")

	api := new(API)
	listener, e := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Printf("Server listening on port %v...\n", port)
	for {
		conn, e := listener.Accept()
		if e != nil {
			panic(e)
		}

		// Set up server side of yamux
		mux, e := yamux.Server(conn, nil)
		if e != nil {
			panic(e)
		}
		muxConn, e := mux.Accept()
		if e != nil {
			panic(e)
		}
		srpc := rpc.NewServer()
		srpc.Register(api)
		go srpc.ServeConn(muxConn)

		// Set up stream back to client
		stream, e := mux.Open()
		if e != nil {
			panic(e)
		}
		crpc := rpc.NewClient(stream)

		// Ask client for player name
		var state common.PlayerState
		e = crpc.Call("API.GetPersonState", 0, &state)
		if e != nil {
			log.Fatal("GetPersonState error:", e)
		}
		p := connectedPerson{state: state, rpc: crpc}
		log.Println(p.state.Name)
		api.connectedPeople = append(api.connectedPeople, &p)
	}
}

type connectedPerson struct {
	rpc   *rpc.Client
	state common.PlayerState
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
