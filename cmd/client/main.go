package main

import (
	"os"
	"strconv"

	// Uncomment for profiling
	// _ "net/http/pprof"

	"github.com/jeffbaumes/gogame/pkg/client"
)

func main() {
	// Uncomment for profiling
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	args := os.Args[1:]
	name := "andrew"
	host := "localhost"
	port := 5555
	var e error
	if len(args) >= 1 {
		name = args[0]
	}
	if len(args) >= 2 {
		host = args[1]
	}
	if len(args) >= 3 {
		port, e = strconv.Atoi(args[2])
		if e != nil {
			panic(e)
		}
	}
	client.Start(name, host, port)
}
