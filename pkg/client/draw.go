package client

import (
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jeffbaumes/gogame/pkg/common"
	"github.com/jeffbaumes/gogame/pkg/scene"
)

func initGlfw() *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(500, 500, "World Blocks", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	return window
}

func initOpenGL() {

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// messageCallback := func(
	// 	source uint32,
	// 	gltype uint32,
	// 	id uint32,
	// 	severity uint32,
	// 	length int32,
	// 	message string,
	// 	userParam unsafe.Pointer,
	// ) {
	// 	s := ""
	// 	if gltype == gl.DEBUG_TYPE_ERROR {
	// 		s = "** ERROR **"
	// 	}
	// 	log.Printf("GL CALLBACK: %s type = 0x%x, severity = 0x%x, message = %s\n", s, gltype, severity, message)
	// }
	// gl.Enable(gl.DEBUG_OUTPUT)
	// gl.DebugMessageCallback(messageCallback, nil)

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func drawFrame(h float32, player *common.Player, text *scene.Text, over *scene.Crosshair, peopleRen *scene.Players, focusRen *scene.FocusCell, bar *scene.Hotbar, health *scene.Health, window *glfw.Window, time float64) {
	universe.Draw(window, time)
	peopleRen.Draw(player, window)
	focusRen.Draw(player, universe.Player.Planet, window)
	over.Draw(window)
	text.Draw(player, window)
	bar.Draw(player, window)
	health.Draw(player, window)

	glfw.PollEvents()
	window.SwapBuffers()
}
