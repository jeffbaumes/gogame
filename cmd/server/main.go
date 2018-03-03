package main

import (
	"github.com/jeffbaumes/gogame/pkg/server"
)

func main() {
	server.Start("default", 0, 5555)
}
