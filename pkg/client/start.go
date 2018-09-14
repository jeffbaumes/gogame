package client

import (
	"fmt"
	"net"
	"net/rpc"
	"runtime"
	"time"

	"github.com/anbcodes/goguigl/gui"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/hashicorp/yamux"
	"github.com/jeffbaumes/buildorb/pkg/common"
	"github.com/jeffbaumes/buildorb/pkg/scene"
)

const (
	targetFPS = 60
	gravity   = 9.8
)

var (
	universe *scene.Universe
	screen   *gui.Screen
	op       *scene.Options
)

// Start starts a client with the given username, host, and port
func Start(username, host string, port int, scr *gui.Screen) {
	screen = scr
	screen.Clear()
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 5555
	}
	window := screen.Window
	if screen.Window == nil {
		runtime.LockOSThread()

		window = initGlfw()
		initOpenGL()
	}

	defer glfw.Terminate()
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

	player := common.NewPlayer(username)
	universe = scene.NewUniverse(player, cRPC)

	planetStates := []*common.PlanetState{}
	e = cRPC.Call("API.GetPlanetStates", 0, &planetStates)
	if e != nil {
		panic(e)
	}
	for _, state := range planetStates {
		planet := common.NewPlanet(*state, cRPC, nil)
		planetRen := scene.NewPlanet(planet)
		universe.AddPlanet(planetRen)
	}

	op = scene.NewOptions(screen)
	player.Planet = universe.PlanetMap[0].Planet
	player.Spawn()

	over := scene.NewCrosshair()
	text := &scene.Text{}
	bar := scene.NewHotbar()
	health := scene.NewHealth()
	player.Mode = "Play"
	// Setup server connection
	smuxConn, e := cmux.Accept()
	if e != nil {
		panic(e)
	}
	s := rpc.NewServer()
	clientAPI := new(API)
	s.Register(clientAPI)
	go s.ServeConn(smuxConn)

	peopleRen := scene.NewPlayers(&universe.ConnectedPeople)
	focusRen := scene.NewFocusCell()

	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	window.SetKeyCallback(keyCallback)
	window.SetCursorPosCallback(cursorPosCallback())
	window.SetSizeCallback(windowSizeCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)

	startTime := time.Now()
	t := startTime
	syncT := t
	for !window.ShouldClose() {
		h := float32(time.Since(t)) / float32(time.Second)
		t = time.Now()
		elapsedSeconds := float64(time.Since(startTime)) / float64(time.Second)

		drawFrame(h, player, text, over, peopleRen, focusRen, bar, health, screen, elapsedSeconds, op)

		player.UpdatePosition(h)

		if float64(time.Since(syncT))/float64(time.Second) > 0.05 {
			syncT = time.Now()
			var ret bool
			cRPC.Go("API.UpdatePersonState", &common.PlayerState{
				Name:     player.Name,
				Position: player.Location(),
				LookDir:  player.LookDir(),
			}, &ret, nil)
		}
		time.Sleep(time.Second/time.Duration(targetFPS) - time.Since(t))
	}
}
