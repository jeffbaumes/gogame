package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jeffbaumes/gogame/pkg/client"
	"github.com/jeffbaumes/gogame/pkg/server"
)

func main() {
	args := os.Args[1:]
	sseed := 1
	sport := 5555
	sworld := "default"
	play := "all"
	name := "andrew"
	host := "localhost"
	port := 5555

	var e error
	if len(args) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter if you want to do server or client or all (put 'all' or 'server' or 'client'): ")

		playstr, _ := reader.ReadString('\n')
		if strings.TrimSpace(playstr) == "client" {
			play = "client"
		}
		if strings.TrimSpace(playstr) == "server" {
			play = "server"
		}
		if play == "server" || play == "all" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter world name for server (leave blank for 'default'): ")

			worldstr, _ := reader.ReadString('\n')
			if strings.TrimSpace(worldstr) != "" {
				sworld = worldstr
			}
			reader = bufio.NewReader(os.Stdin)
			fmt.Print("Enter seed for server (leave blank for 1): ")
			seedstr, _ := reader.ReadString('\n')
			if strings.TrimSpace(seedstr) != "" {
				sseed, e = strconv.Atoi(strings.TrimSpace(seedstr))
				if e != nil {
					panic(e)
				}
			}
			reader = bufio.NewReader(os.Stdin)
			fmt.Print("Enter port for server (leave blank for 5555): ")
			portstr, _ := reader.ReadString('\n')
			if strings.TrimSpace(portstr) != "" {
				sport, e = strconv.Atoi(strings.TrimSpace(portstr))
				if e != nil {
					panic(e)
				}
			}
		}
	} else {
		play = args[6]
		sworld = args[0]
		sport, e = strconv.Atoi(strings.TrimSpace(args[1]))
		if e != nil {
			panic(e)
		}
		sseed, e = strconv.Atoi(strings.TrimSpace(args[2]))
		if e != nil {
			panic(e)
		}
		name = args[3]
		host = args[4]
		port, e = strconv.Atoi(strings.TrimSpace(args[5]))
		if e != nil {
			panic(e)
		}
	}
	if len(args) == 0 {
		if play == "client" || play == "all" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter your name DO NOT LEAVE BLANK: ")

			namestr, _ := reader.ReadString('\n')
			if strings.TrimSpace(namestr) != "" {
				name = namestr
			}
			reader = bufio.NewReader(os.Stdin)
			fmt.Print("Enter host for client (leave blank for 'localhost'): ")
			hoststr, _ := reader.ReadString('\n')
			if strings.TrimSpace(hoststr) != "" {
				host = strings.TrimSpace(hoststr)
			}
			reader = bufio.NewReader(os.Stdin)
			fmt.Print("Enter port for client (leave blank for 5555): ")
			portstr, _ := reader.ReadString('\n')
			if strings.TrimSpace(portstr) != "" {
				port, e = strconv.Atoi(strings.TrimSpace(portstr))
				if e != nil {
					panic(e)
				}
			}
		}
	}
	if play == "server" || play == "all" {
		if play == "all" {
			go server.Start(sworld, sseed, sport)
		} else {
			server.Start(sworld, sseed, sport)
		}
	}
	time.Sleep(1)
	if play == "client" || play == "all" {
		client.Start(name, host, port)
	}
}
