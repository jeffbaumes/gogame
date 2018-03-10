package main

import (
	"os"

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
	if len(args) >= 1 {
		name = args[0]
	}
	if len(args) >= 2 {
		host = args[1]
	}
	client.Start(name, host, 5555)
}
