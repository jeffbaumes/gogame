package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	// Uncomment for profiling
	// _ "net/http/pprof"

	"github.com/jeffbaumes/buildorb/pkg/client"
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
	if len(args) == 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your name DO NOT LEAVE BLANK: ")

		namestr, _ := reader.ReadString('\n')
		if strings.TrimSpace(namestr) != "" {
			name = namestr
		}
		reader = bufio.NewReader(os.Stdin)
		fmt.Print("Enter host (leave blank for 'localhost'): ")
		hoststr, _ := reader.ReadString('\n')
		if strings.TrimSpace(hoststr) != "" {
			host = strings.TrimSpace(hoststr)
		}
		reader = bufio.NewReader(os.Stdin)
		fmt.Print("Enter port (leave blank for 5555): ")
		portstr, _ := reader.ReadString('\n')
		if strings.TrimSpace(portstr) != "" {
			port, e = strconv.Atoi(strings.TrimSpace(portstr))
			if e != nil {
				panic(e)
			}
		}
	}
	client.Start(name, host, port, nil)
}
