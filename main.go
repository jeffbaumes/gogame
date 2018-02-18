package main

import (
	"log"
	"math"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	opensimplex "github.com/ojrac/opensimplex-go"
)

const (
	fps = 60
)

var (
	noise         = opensimplex.NewWithSeed(0)
	cursorGrabbed = false
	player        = newPerson()
	aspectRatio   = float32(1.0)
	lastCursor    = mgl64.Vec2{math.NaN(), math.NaN()}
	g             = 9.8
)

func cursorPosCallback(w *glfw.Window, xpos float64, ypos float64) {
	if !cursorGrabbed {
		lastCursor[0] = math.NaN()
		lastCursor[1] = math.NaN()
		return
	}
	curCursor := mgl64.Vec2{xpos, ypos}
	if math.IsNaN(lastCursor[0]) {
		lastCursor = curCursor
	}
	delta := curCursor.Sub(lastCursor)
	lookHeadingDelta := -0.1 * delta[0]
	normalDir := player.loc.Normalize()
	player.lookHeading = mgl32.QuatRotate(float32(lookHeadingDelta*math.Pi/180.0), normalDir).Rotate(player.lookHeading)
	player.lookAltitude = player.lookAltitude - 0.1*delta[1]
	player.lookAltitude = math.Max(math.Min(player.lookAltitude, 89.9), -89.9)
	lastCursor = curCursor
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if !cursorGrabbed {
		return
	}
	switch action {
	case glfw.Press:
		switch key {
		case glfw.KeySpace:
			player.altVel = player.walkVel
		case glfw.KeyLeftShift:
			player.altVel = -player.walkVel
		case glfw.KeyW:
			player.forwardVel = player.walkVel
		case glfw.KeyS:
			player.forwardVel = -player.walkVel
		case glfw.KeyEscape:
			w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			cursorGrabbed = false
		}
	case glfw.Release:
		switch key {
		case glfw.KeySpace:
			player.altVel = 0
		case glfw.KeyLeftShift:
			player.altVel = 0
		case glfw.KeyW:
			player.forwardVel = 0
		case glfw.KeyS:
			player.forwardVel = 0
		}
	}
}

func windowSizeCallback(w *glfw.Window, width, height int) {
	aspectRatio = float32(width) / float32(height)
}

func mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if !cursorGrabbed && action == glfw.Press {
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		cursorGrabbed = true
	}
}

func main() {
	runtime.LockOSThread()

	window := InitGlfw()
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	cursorGrabbed = true
	window.SetKeyCallback(keyCallback)
	window.SetCursorPosCallback(cursorPosCallback)
	window.SetSizeCallback(windowSizeCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)
	defer glfw.Terminate()
	program := InitOpenGL()
	projection := UniformLocation(program, "proj")

	p := newPlanet(20.0, 70.0, 80, 30, 20, 15, 20)
	log.Println(p.cartesianToCell(mgl32.Vec3{19.5, 0, 0}))
	t := time.Now()
	for !window.ShouldClose() {
		h := float32(time.Since(t)) / float32(time.Second)
		t = time.Now()

		draw(h, p, window, program, projection)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}
}

func draw(h float32, p *planet, window *glfw.Window, program uint32, projection int32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	normalDir := player.loc.Normalize()

	// To project vector to plane, subtract vector projected to normal
	forward := player.lookHeading.Sub(normalDir.Mul(player.lookHeading.Dot(normalDir))).Normalize()

	lookDir := mgl32.QuatRotate(float32((player.lookAltitude-90.0)*math.Pi/180.0), forward).Rotate(normalDir)
	view := mgl32.LookAtV(player.loc, player.loc.Add(lookDir), player.loc.Normalize())
	perspective := mgl32.Perspective(45, aspectRatio, 0.01, 100)
	proj := perspective.Mul4(view)
	gl.UniformMatrix4fv(projection, 1, false, &proj[0])

	p.draw()

	// Update position
	if cursorGrabbed {
		player.loc = player.loc.Add(lookDir.Mul(player.forwardVel * h))
		player.loc = player.loc.Add(normalDir.Mul(player.altVel * h))
	}

	glfw.PollEvents()
	window.SwapBuffers()
}
