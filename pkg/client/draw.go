package client

import (
	"log"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/gogame/pkg/client/scene"
	"github.com/jeffbaumes/gogame/pkg/common"
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
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	gl.Enable(gl.DEPTH_TEST)
	// gl.Enable(gl.POLYGON_OFFSET_FILL)
	// gl.PolygonOffset(2, 0)

	gl.LineWidth(5)
	gl.Enable(gl.LINE_SMOOTH)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func drawFrame(h float32, player *common.Player, text *scene.Text, over *scene.Crosshair, planetRen *scene.Planet, peopleRen *scene.Players, focusRen *scene.FocusCell, bar *scene.Hotbar, window *glfw.Window, timeOfDay float64) {
	sunAngle := timeOfDay * math.Pi * 2
	sunDir := mgl32.Vec3{float32(math.Sin(sunAngle)), float32(math.Cos(sunAngle)), 0}
	vpnDotSun := float64(player.Loc.Normalize().Dot(sunDir))
	light1Color := mgl32.Vec3{0.5, 0.7, 1.0}
	light1 := math.Max(math.Sqrt(vpnDotSun), 0)
	if math.IsNaN(light1) {
		light1 = 0
	}
	light2Color := mgl32.Vec3{0, 0, 0}
	light2 := math.Max(math.Sqrt(1-vpnDotSun), 0)
	if math.IsNaN(light2) {
		light2 = 0
	}
	light3Color := mgl32.Vec3{0.7, 0.5, 0.4}
	light3 := math.Max(0.6-math.Sqrt(math.Abs(vpnDotSun)), 0)
	if math.IsNaN(light3) {
		light3 = 0
	}
	light := light1Color.Mul(float32(light1)).Add(light2Color.Mul(float32(light2))).Add(light3Color.Mul(float32(light3)))

	gl.ClearColor(light.X(), light.Y(), light.Z(), 1)
	planetRen.Planet.CartesianToCellIndex(player.Loc)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	planetRen.Draw(player, window, timeOfDay)
	peopleRen.Draw(player, window)
	focusRen.Draw(player, planetRen.Planet, window)
	over.Draw(window)
	text.Draw(player, window)
	bar.Draw(player, window)

	glfw.PollEvents()
	window.SwapBuffers()
}
