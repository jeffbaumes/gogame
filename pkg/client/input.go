package client

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/jeffbaumes/gogame/pkg/geom"
)

func cursorGrabbed(w *glfw.Window) bool {
	return w.GetInputMode(glfw.CursorMode) == glfw.CursorDisabled
}

func framebufferSize(w *glfw.Window) (fbw, fbh int) {
	fbw, fbh = w.GetFramebufferSize()
	return
}

func cursorPosCallback(player *person) func(w *glfw.Window, xpos, ypos float64) {
	lastCursor := mgl64.Vec2{math.NaN(), math.NaN()}
	return func(w *glfw.Window, xpos, ypos float64) {
		if !cursorGrabbed(w) {
			lastCursor = mgl64.Vec2{math.NaN(), math.NaN()}
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
}

func glToPixel(w *glfw.Window, xpos, ypos float64) (xpix, ypix float64) {
	winw, winh := w.GetSize()
	return float64(winw) * (xpos + 1) / 2, float64(winh) * (-ypos + 1) / 2
}

func keyCallback(player *person) func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	return func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if player.inInventory {
			slot := -1
			switch action {
			case glfw.Press:
				switch key {
				case glfw.Key1:
					slot = 0
				case glfw.Key2:
					slot = 1
				case glfw.Key3:
					slot = 2
				case glfw.Key4:
					slot = 3
				case glfw.Key5:
					slot = 4
				case glfw.Key6:
					slot = 5
				case glfw.Key7:
					slot = 6
				case glfw.Key8:
					slot = 7
				case glfw.Key9:
					slot = 8
				case glfw.Key0:
					slot = 9
				case glfw.KeyMinus:
					slot = 10
				case glfw.KeyEqual:
					slot = 11
				}
			}
			if slot >= 0 {
				xpos, ypos := w.GetCursorPos()
				winw, winh := w.GetSize()
				aspect := float32(winw) / float32(winh)
				println(int(xpos), int(ypos))
				for m := range geom.Materials {
					sz := float32(0.05)
					px := 1.25 * 2 * sz * (float32(m) - float32(len(geom.Materials))/2)
					py := 1 - 0.25*aspect
					scale := sz
					xMin, yMin := glToPixel(w, float64(px-scale), float64(py+scale*aspect))
					xMax, yMax := glToPixel(w, float64(px+scale), float64(py-scale*aspect))
					println(int(xMin), int(xMax), int(yMin), int(yMax))
					if float64(xpos) >= xMin && float64(xpos) <= xMax && float64(ypos) >= yMin && float64(ypos) <= yMax {
						player.hotBar[slot] = m
					}
				}
			}
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
			case glfw.Key1:
				player.activeHotBarSlot = 0
			case glfw.Key2:
				player.activeHotBarSlot = 1
			case glfw.Key3:
				player.activeHotBarSlot = 2
			case glfw.Key4:
				player.activeHotBarSlot = 3
			case glfw.Key5:
				player.activeHotBarSlot = 4
			case glfw.Key6:
				player.activeHotBarSlot = 5
			case glfw.Key7:
				player.activeHotBarSlot = 6
			case glfw.Key8:
				player.activeHotBarSlot = 7
			case glfw.Key9:
				player.activeHotBarSlot = 8
			case glfw.Key0:
				player.activeHotBarSlot = 9
			case glfw.KeyMinus:
				player.activeHotBarSlot = 10
			case glfw.KeyEqual:
				player.activeHotBarSlot = 11
			case glfw.KeyE:
				if player.inInventory == false {
					player.hotBarOn = true
					player.inInventory = true
					w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
				} else {
					player.inInventory = false
					w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
				}
			case glfw.KeyH:
				if player.hotBarOn == false {
					player.hotBarOn = true
				} else {
					player.hotBarOn = false
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
				if player.gameMode == flying {
					player.fallVel = 0
				}
			case glfw.KeyEscape:
				w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
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
}

func windowSizeCallback(w *glfw.Window, wd, ht int) {
	fwidth, fheight := framebufferSize(w)
	gl.Viewport(0, 0, int32(fwidth), int32(fheight))
}

func mouseButtonCallback(player *person, planetRen *planetRenderer) func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	return func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		planet := planetRen.planet
		if cursorGrabbed(w) {
			if action == glfw.Press && button == glfw.MouseButtonLeft {
				increment := player.lookDir().Mul(0.05)
				pos := player.loc
				for i := 0; i < 100; i++ {
					pos = pos.Add(increment)
					cell := planet.CartesianToCell(pos)
					if cell != nil && cell.Material != geom.Air {
						cellIndex := planet.CartesianToCellIndex(pos)
						planetRen.setCellMaterial(cellIndex, geom.Air)
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
					nextCellIndex := planet.CartesianToCellIndex(pos)
					if nextCellIndex != cellIndex {
						prevCellIndex = cellIndex
						cellIndex = nextCellIndex
						cell := planet.CellIndexToCell(cellIndex)
						if cell != nil && cell.Material != geom.Air {
							if prevCellIndex.Lon != -1 {
								planetRen.setCellMaterial(prevCellIndex, player.hotBar[player.activeHotBarSlot])
							}
							break
						}
					}
				}
			}
		} else {
			if action == glfw.Press {
				w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
			}
		}
	}
}
