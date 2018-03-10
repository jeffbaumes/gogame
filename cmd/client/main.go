package main

import (
	"os"

	"github.com/jeffbaumes/gogame/pkg/client"
)

func main() {
	args := os.Args[1:]
	name := "andrew"
	if len(args) >= 1 {
		name = args[0]
	}
	client.Start(name, "localhost", 5555)
}
