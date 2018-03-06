package client

import (
	"fmt"
	"math"
	"net"
	"net/rpc"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/hashicorp/yamux"
	"github.com/jeffbaumes/gogame/pkg/geom"
)

const (
	fps = 60
)

const (
	normal       = iota
	flying       = iota
	numGameModes = iota
)

var (
	cursorGrabbed  = false
	renderDistance = 4
	player         = newPerson()
	p              *geom.Planet
	aspectRatio    = float32(1.0)
	lastCursor     = mgl64.Vec2{math.NaN(), math.NaN()}
	g              = 9.8
	cRPC           *rpc.Client
)

// Start starts a client with the given username, host, and port
func Start(username, host string, port int) {
	runtime.LockOSThread()

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
	cRPC = rpc.NewClient(stream)

	// Setup server connection
	smuxConn, e := cmux.Accept()
	if e != nil {
		panic(e)
	}
	s := rpc.NewServer()
	clientAPI := new(Client)
	s.Register(clientAPI)
	go s.ServeConn(smuxConn)

	window := initGlfw()
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	cursorGrabbed = true
	window.SetKeyCallback(keyCallback)
	window.SetCursorPosCallback(cursorPosCallback)
	window.SetSizeCallback(windowSizeCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)
	defer glfw.Terminate()
	program, hudProgram, textProgram := initOpenGL()
	initHUD()
	initText()
	projection := uniformLocation(program, "proj")
	hudProjection := uniformLocation(hudProgram, "proj")
	textTexture := uniformLocation(textProgram, "texture")

	p = geom.NewPlanet(50.0, 16, 0, cRPC)
	t := time.Now()
	for !window.ShouldClose() {
		h := float32(time.Since(t)) / float32(time.Second)
		t = time.Now()

		drawFrame(h, window, program, hudProgram, textProgram, projection, hudProjection, textTexture)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}
}

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
			if player.gameMode == normal {
				player.holdingJump = true
			} else {
				player.upVel = player.walkVel
			}
		case glfw.KeyLeftShift:
			if player.gameMode == flying {
				player.downVel = player.walkVel
			}
		case glfw.KeyW:
			player.forwardVel = player.walkVel
		case glfw.KeyS:
			player.backVel = player.walkVel
		case glfw.KeyD:
			player.rightVel = player.walkVel
		case glfw.KeyA:
			player.leftVel = player.walkVel
		case glfw.KeyM:
			player.gameMode++
			if player.gameMode >= numGameModes {
				player.gameMode = 0
			}
		case glfw.KeyEscape:
			w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			cursorGrabbed = false
		}
	case glfw.Release:
		switch key {
		case glfw.KeySpace:
			player.holdingJump = false
			player.upVel = 0
		case glfw.KeyLeftShift:
			player.downVel = 0
		case glfw.KeyW:
			player.forwardVel = 0
		case glfw.KeyS:
			player.backVel = 0
		case glfw.KeyD:
			player.rightVel = 0
		case glfw.KeyA:
			player.leftVel = 0
		}
	}
}

func windowSizeCallback(w *glfw.Window, wd, ht int) {
	aspectRatio = float32(wd) / float32(ht)
	fwidth, fheight := w.GetFramebufferSize()
	gl.Viewport(0, 0, int32(fwidth), int32(fheight))
	width = wd
	height = ht
}

func mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if cursorGrabbed {
		if action == glfw.Press && button == glfw.MouseButtonLeft {
			increment := player.lookDir().Mul(0.05)
			pos := player.loc
			for i := 0; i < 100; i++ {
				pos = pos.Add(increment)
				cell := p.CartesianToCell(pos)
				if cell != nil && cell.Material != geom.Air {
					cellIndex := p.CartesianToCellIndex(pos)
					p.SetCellMaterial(cellIndex, geom.Air)
					break
				}
			}
		} else if action == glfw.Press && button == glfw.MouseButtonRight {
			increment := player.lookDir().Mul(0.05)
			pos := player.loc
			prevCellIndex := geom.CellIndex{Lon: -1, Lat: -1, Alt: -1}
			cellIndex := geom.CellIndex{}
			for i := 0; i < 100; i++ {
				pos = pos.Add(increment)
				nextCellIndex := p.CartesianToCellIndex(pos)
				if nextCellIndex != cellIndex {
					prevCellIndex = cellIndex
					cellIndex = nextCellIndex
					cell := p.CellIndexToCell(cellIndex)
					if cell != nil && cell.Material != geom.Air {
						if prevCellIndex.Lon != -1 {
							p.SetCellMaterial(prevCellIndex, geom.Rock)
						}
						break
					}
				}
			}
		}
	} else {
		if action == glfw.Press {
			w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
			cursorGrabbed = true
		}
	}
}

func project(a mgl32.Vec3, b mgl32.Vec3) mgl32.Vec3 {
	bn := b.Normalize()
	return bn.Mul(a.Dot(bn))
}

func projectToPlane(v mgl32.Vec3, n mgl32.Vec3) mgl32.Vec3 {
	if v[0] == 0 && v[1] == 0 && v[2] == 0 {
		return v
	}
	// To project vector to plane, subtract vector projected to normal
	return v.Sub(project(v, n))
}

func min(val, a int) int {
	if val < a {
		return val
	}
	return a
}

func max(val, a int) int {
	if val > a {
		return val
	}
	return a
}

func drawFrame(h float32, window *glfw.Window, program, hudProgram, textProgram uint32, projection, hudProjection, textTexture int32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(program)
	lookDir := player.lookDir()
	view := mgl32.LookAtV(player.loc, player.loc.Add(lookDir), player.loc.Normalize())
	perspective := mgl32.Perspective(45, aspectRatio, 0.01, 1000)
	proj := perspective.Mul4(view)
	gl.UniformMatrix4fv(projection, 1, false, &proj[0])
	drawPlanet(p)

	gl.UseProgram(hudProgram)
	projHUD := mgl32.Scale3D(1/float32(width), 1/float32(height), 1.0)
	gl.UniformMatrix4fv(hudProjection, 1, false, &projHUD[0])
	drawHUD()

	r, theta, phi := mgl32.CartesianToSpherical(player.loc)
	text.statusLine.x = 1
	text.statusLine.y = 1
	text.statusLine.str = fmt.Sprintf("LAT %v, LON %v, ALT %v", int(theta/math.Pi*180-90+0.5), int(phi/math.Pi*180+0.5), int(r+0.5))
	initTextGeom()

	gl.UseProgram(textProgram)
	gl.Uniform1i(textTexture, int32(textTextureValue))
	drawText()

	if !cursorGrabbed {
		glfw.PollEvents()
		window.SwapBuffers()
		return
	}

	// Update position
	up := player.loc.Normalize()
	right := player.lookHeading.Cross(up)
	if player.gameMode == normal {
		feet := player.loc.Sub(up.Mul(float32(player.height)))
		feetCell := p.CartesianToCell(feet)

		ind := p.CartesianToChunkIndex(feet)

		for lon := ind.Lon - renderDistance; lon <= ind.Lon+renderDistance; lon++ {
			validLon := lon
			for validLon < 0 {
				validLon += p.LonCells / geom.ChunkSize
			}
			for validLon >= p.LonCells/geom.ChunkSize {
				validLon -= p.LonCells / geom.ChunkSize
			}
			latMin := max(ind.Lat-renderDistance, 0)
			latMax := min(ind.Lat+renderDistance, p.LatCells/geom.ChunkSize-1)
			for lat := latMin; lat <= latMax; lat++ {
				for alt := 0; alt < p.AltCells/geom.ChunkSize; alt++ {
					p.GetChunk(geom.ChunkIndex{Lon: validLon, Lat: lat, Alt: alt})
				}
			}
		}

		falling := feetCell == nil || feetCell.Material == geom.Air
		if falling {
			player.fallVel -= 20 * h
		} else if player.holdingJump && !player.inJump {
			player.fallVel = 7
			player.inJump = true
		} else {
			player.fallVel = 0
			player.inJump = false
		}

		playerVel := mgl32.Vec3{}
		playerVel = playerVel.Add(up.Mul(player.fallVel))
		playerVel = playerVel.Add(player.lookHeading.Mul((player.forwardVel - player.backVel)))
		playerVel = playerVel.Add(right.Mul((player.rightVel - player.leftVel)))

		player.loc = player.loc.Add(playerVel.Mul(h))
		for height := p.AltDelta / 2; height < player.height; height += p.AltDelta {
			player.collide(p, float32(height), geom.CellLoc{Lon: 0, Lat: 0, Alt: -1})
			player.collide(p, float32(height), geom.CellLoc{Lon: 1, Lat: 0, Alt: 0})
			player.collide(p, float32(height), geom.CellLoc{Lon: -1, Lat: 0, Alt: 0})
			player.collide(p, float32(height), geom.CellLoc{Lon: 0, Lat: 1, Alt: 0})
			player.collide(p, float32(height), geom.CellLoc{Lon: 0, Lat: -1, Alt: 0})
		}
	} else if player.gameMode == flying {
		player.loc = player.loc.Add(up.Mul((player.upVel - player.downVel) * h))
		player.loc = player.loc.Add(lookDir.Mul((player.forwardVel - player.backVel) * h))
		player.loc = player.loc.Add(right.Mul((player.rightVel - player.leftVel) * h))
	}

	glfw.PollEvents()
	window.SwapBuffers()
}
