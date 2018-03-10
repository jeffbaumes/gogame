package client

import (
	"fmt"
	"net"
	"net/rpc"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/hashicorp/yamux"
	"github.com/jeffbaumes/gogame/pkg/geom"
)

const (
	targetFPS      = 60
	renderDistance = 4
	gravity        = 9.8
)

// Start starts a client with the given username, host, and port
func Start(username, host string, port int) {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()
	initOpenGL()

	player := newPerson(username)

	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", host, port))
	if err != nil {
		panic(err)
	}

	// Setup client side of yamux
	cmux, e := yamux.Client(conn, nil)
	if e != nil {
		panic(e)
	}
	stream, e := cmux.Open()
	if e != nil {
		panic(e)
	}
	cRPC := rpc.NewClient(stream)

	// Create planet
	planet := geom.NewPlanet(50.0, 16, 0, cRPC)
	planetRen := newPlanetRenderer(planet)
	over := newOverlay()
	text := newScreenText()

	// Setup server connection
	smuxConn, e := cmux.Accept()
	if e != nil {
		panic(e)
	}
	s := rpc.NewServer()
	clientAPI := newAPI(planetRen, player)
	s.Register(clientAPI)
	go s.ServeConn(smuxConn)

	peopleRen := newPeopleRenderer(clientAPI)

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetKeyCallback(keyCallback(player))
	window.SetCursorPosCallback(cursorPosCallback(player))
	window.SetSizeCallback(windowSizeCallback)
	window.SetMouseButtonCallback(mouseButtonCallback(player, planetRen))

	t := time.Now()
	syncT := t
	for !window.ShouldClose() {
		h := float32(time.Since(t)) / float32(time.Second)
		t = time.Now()

		drawFrame(h, player, text, over, planetRen, peopleRen, window)

		if cursorGrabbed(window) {
			player.updatePosition(h, planet)
		}

		if float64(time.Since(syncT))/float64(time.Second) > 0.05 {
			syncT = time.Now()
			var ret bool
			cRPC.Call("API.UpdatePersonState", &geom.PersonState{
				Name:     player.name,
				Position: player.loc,
				LookDir:  player.lookDir(),
			}, &ret)
		}

		time.Sleep(time.Second/time.Duration(targetFPS) - time.Since(t))
	}
}
