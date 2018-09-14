package server

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"

	"github.com/hashicorp/yamux"
	"github.com/jeffbaumes/buildorb/pkg/common"
	_ "github.com/mattn/go-sqlite3" // Needed to use sqlite
)

var (
	universe *common.Universe
)

type server struct {
	system string
}

func writefile(t string) {

	err := ioutil.WriteFile("server.buildorb", []byte(t), 0644)
	if err != nil {
		panic(err)
	}
}
func readfile() string {

	// read the whole file at once
	b, err := ioutil.ReadFile("server.buildorb")
	if err != nil {
		os.Create("server.buildorb")
		return readfile()
	}
	// write the whole body at once
	// err := ioutil.WriteFile("profils.txt", []byte(t), 0644)
	// if err != nil {
	// 	panic(err)
	// }
	return string(b)
}

func getsystem() (f5 string) {
	f2 := readfile()
	f := strings.Split(f2, ";")
	for _, f3 := range f {
		f4 := strings.Split(f3, "=")
		if f4[0] == "system" {
			f5 = f4[1]
		}
	}
	return
}

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

	universe = common.NewUniverse(db, getsystem())

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
