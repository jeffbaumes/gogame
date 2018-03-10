package main

import (
	"os"

	"github.com/jeffbaumes/gogame/pkg/client"
)

func main() {
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
