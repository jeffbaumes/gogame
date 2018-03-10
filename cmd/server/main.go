package main

import (
	"os"
	"strconv"

	"github.com/jeffbaumes/gogame/pkg/server"
)

func main() {
	args := os.Args[1:]
	seed := 1
	port := 5555
	var e error
	if len(args) >= 1 {
		seed, e = strconv.Atoi(args[0])
		if e != nil {
			panic(e)
		}
	}
	if len(args) >= 2 {
		port, e = strconv.Atoi(args[1])
		if e != nil {
			panic(e)
		}
	}
	server.Start("default", seed, port)
}
