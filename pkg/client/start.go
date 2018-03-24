package client

import (
	"fmt"
	"math"
	"net"
	"net/rpc"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/hashicorp/yamux"
	"github.com/jeffbaumes/gogame/pkg/client/scene"
	"github.com/jeffbaumes/gogame/pkg/common"
)

const (
	targetFPS     = 60
	gravity       = 9.8
	secondsPerDay = 300
)

// Start starts a client with the given username, host, and port
func Start(username, host string, port int) {
	runtime.LockOSThread()

	window := initGlfw()
	defer glfw.Terminate()
	initOpenGL()

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
	planetID := 0
	planetState := common.PlanetState{}
	e = cRPC.Call("API.GetPlanetState", planetID, &planetState)
	if e != nil {
		panic(e)
	}

	planet := common.NewPlanet(planetState, cRPC, nil)
	player := common.NewPlayer(username, planet)
	planetRen := scene.NewPlanet(planet)
	over := scene.NewCrosshair()
	text := scene.NewText()
	bar := scene.NewHotbar()
	health := scene.NewHealth()

	// Setup server connection
	smuxConn, e := cmux.Accept()
	if e != nil {
		panic(e)
	}
	s := rpc.NewServer()
	clientAPI := newAPI(planetRen, player)
	s.Register(clientAPI)
	go s.ServeConn(smuxConn)

	peopleRen := scene.NewPlayers(&clientAPI.connectedPeople)
	focusRen := scene.NewFocusCell()

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetKeyCallback(keyCallback(player, cRPC))
	window.SetCursorPosCallback(cursorPosCallback(player))
	window.SetSizeCallback(windowSizeCallback)
	window.SetMouseButtonCallback(mouseButtonCallback(player, planetRen, &clientAPI.connectedPeople, cRPC))

	startTime := time.Now()
	t := startTime
	syncT := t
	for !window.ShouldClose() {
		h := float32(time.Since(t)) / float32(time.Second)
		t = time.Now()

		elapsedSeconds := float64(time.Since(startTime)) / float64(time.Second)
		_, timeOfDay := math.Modf(elapsedSeconds / secondsPerDay)

		drawFrame(h, player, text, over, planetRen, peopleRen, focusRen, bar, health, window, timeOfDay)

		player.UpdatePosition(h, planet)

		if float64(time.Since(syncT))/float64(time.Second) > 0.05 {
			syncT = time.Now()
			var ret bool
			cRPC.Go("API.UpdatePersonState", &common.PlayerState{
				Name:     player.Name,
				Position: player.Loc,
				LookDir:  player.LookDir(),
			}, &ret, nil)
		}

		time.Sleep(time.Second/time.Duration(targetFPS) - time.Since(t))
	}
}
