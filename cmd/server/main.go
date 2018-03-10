package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jeffbaumes/gogame/pkg/server"
)

func main() {
	args := os.Args[1:]
	seed := 1
	port := 5555
	world := "default"
	var e error
	if len(args) >= 1 {
		world = args[0]
	}
	if len(args) >= 2 {
		seed, e = strconv.Atoi(args[1])
		if e != nil {
			panic(e)
		}
	}
	if len(args) >= 3 {
		port, e = strconv.Atoi(args[2])
		if e != nil {
			panic(e)
		}
	}
	if len(args) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter world name (leave blank for 'default'): ")

		worldstr, _ := reader.ReadString('\n')
		if worldstr != "" {
			world = worldstr
		}
		reader = bufio.NewReader(os.Stdin)
		fmt.Print("Enter seed (leave blank for 1): ")
		seedstr, _ := reader.ReadString('\n')
		if seedstr != "" {
			seed, e = strconv.Atoi(strings.TrimSpace(seedstr))
			if e != nil {
				panic(e)
			}
		}
		reader = bufio.NewReader(os.Stdin)
		fmt.Print("Enter port (leave blank for 5555): ")
		portstr, _ := reader.ReadString('\n')
		if portstr != "" {
			port, e = strconv.Atoi(strings.TrimSpace(portstr))
			if e != nil {
				panic(e)
			}
		}
	}
	server.Start(world, seed, port)
}
