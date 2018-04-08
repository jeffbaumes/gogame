package client

import (
	"fmt"
	"log"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/jeffbaumes/gogame/pkg/common"
	"github.com/jeffbaumes/gogame/pkg/scene"
)

func cursorGrabbed(w *glfw.Window) bool {
	return w.GetInputMode(glfw.CursorMode) == glfw.CursorDisabled
}

func cursorPosCallback() func(w *glfw.Window, xpos, ypos float64) {
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
		universe.Player.Swivel(delta[0], delta[1])
		lastCursor = curCursor
	}
}

func glToPixel(w *glfw.Window, xpos, ypos float64) (xpix, ypix float64) {
	winw, winh := w.GetSize()
	return float64(winw) * (xpos + 1) / 2, float64(winh) * (-ypos + 1) / 2
}

func sendText(text string) {
	if text == "enter" {
		universe.Player.Intext = false
		// NewText()
		var ret bool
		universe.RPC.Go("API.SendText", universe.Player.Text, &ret, nil)
		universe.Player.Text = ""
		// need to send text to server
	} else if text == "delete" {
		if len(universe.Player.Text) != 0 {
			universe.Player.Text = universe.Player.Text[0:(len(universe.Player.Text) - 1)]
		}
	} else {
		universe.Player.Text = universe.Player.Text + text
	}
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	player := universe.Player
	if action == glfw.Press && key == glfw.KeyO {
		player.Spawn()
	}
	if player.InInventory {
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
			for m := range common.Materials {
				sz := float32(0.03)
				px := 1.25 * 2 * sz * (float32(m) - float32(len(common.Materials))/2)
				py := 1 - 0.25*aspect
				scale := sz
				xMin, yMin := glToPixel(w, float64(px-scale), float64(py+scale*aspect))
				xMax, yMax := glToPixel(w, float64(px+scale), float64(py-scale*aspect))
				if float64(xpos) >= xMin && float64(xpos) <= xMax && float64(ypos) >= yMin && float64(ypos) <= yMax {
					player.Hotbar[slot] = m
				}
			}
		}
	}
	if player.Intext == false {
		switch action {
		case glfw.Press:
			switch key {
			case glfw.KeySpace:
				if cursorGrabbed(w) {
					if player.GameMode == common.Normal {
						player.HoldingJump = true
					} else {
						player.UpVel = player.WalkVel
					}
				}
			case glfw.KeyLeftShift:
				if cursorGrabbed(w) && player.GameMode == common.Flying {
					player.DownVel = player.WalkVel
				}
			case glfw.Key1:
				player.ActiveHotBarSlot = 0
			case glfw.Key2:
				player.ActiveHotBarSlot = 1
			case glfw.Key3:
				player.ActiveHotBarSlot = 2
			case glfw.Key4:
				player.ActiveHotBarSlot = 3
			case glfw.Key5:
				player.ActiveHotBarSlot = 4
			case glfw.Key6:
				player.ActiveHotBarSlot = 5
			case glfw.Key7:
				player.ActiveHotBarSlot = 6
			case glfw.Key8:
				player.ActiveHotBarSlot = 7
			case glfw.Key9:
				player.ActiveHotBarSlot = 8
			case glfw.Key0:
				player.ActiveHotBarSlot = 9
			case glfw.KeyMinus:
				player.ActiveHotBarSlot = 10
			case glfw.KeyEqual:
				player.ActiveHotBarSlot = 11
			case glfw.KeyE:
				if player.InInventory == false {
					player.HotbarOn = true
					player.InInventory = true
					w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
				} else {
					player.InInventory = false
					w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
				}
			case glfw.KeyT:
				if player.Intext == true {
					player.Intext = false
				} else {
					player.Intext = true
				}
			case glfw.KeyH:
				if player.HotbarOn == false {
					player.HotbarOn = true
				} else {
					player.HotbarOn = false
				}
			case glfw.KeyW:
				player.ForwardVel = player.WalkVel
			case glfw.KeyS:
				player.BackVel = player.WalkVel
			case glfw.KeyD:
				player.RightVel = player.WalkVel
			case glfw.KeyA:
				player.LeftVel = player.WalkVel
			case glfw.KeyM:
				player.GameMode++
				if player.GameMode >= common.NumGameModes {
					player.GameMode = 0
				}
				if player.GameMode == common.Flying {
					player.FallVel = 0
				}
			case glfw.KeyK:
				player.Apex = !player.Apex
			case glfw.KeyP:
				player.Planet = universe.PlanetMap[0].Planet
				player.Spawn()
			case glfw.KeyRightBracket:
				id := player.Planet.ID + 1
				if universe.PlanetMap[id] == nil {
					id = 0
				}
				player.Planet = universe.PlanetMap[id].Planet
				player.Spawn()
			case glfw.KeyLeftBracket:
				id := player.Planet.ID - 1
				if universe.PlanetMap[id] == nil {
					if id < 0 {
						for i := range universe.PlanetMap {
							if i > id {
								id = i
							}
						}
					}
				}
				player.Planet = universe.PlanetMap[id].Planet
				player.Spawn()
			case glfw.KeyEscape:
				w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			}
		case glfw.Release:
			switch key {
			case glfw.KeySpace:
				player.HoldingJump = false
				player.UpVel = 0
			case glfw.KeyLeftShift:
				player.DownVel = 0
			case glfw.KeyW:
				player.ForwardVel = 0
			case glfw.KeyS:
				player.BackVel = 0
			case glfw.KeyD:
				player.RightVel = 0
			case glfw.KeyA:
				player.LeftVel = 0
			}
		}
	} else {
		switch action {
		case glfw.Press:
			switch key {
			case glfw.KeyEscape:
				player.Intext = false
			case glfw.Key1:
				sendText("1")
			case glfw.Key2:
				sendText("2")
			case glfw.Key3:
				sendText("3")
			case glfw.Key4:
				sendText("4")
			case glfw.Key5:
				sendText("5")
			case glfw.Key6:
				sendText("6")
			case glfw.Key7:
				sendText("7")
			case glfw.Key8:
				sendText("8")
			case glfw.Key9:
				sendText("9")
			case glfw.Key0:
				sendText("0")
			case glfw.KeyQ:
				sendText("q")
			case glfw.KeyW:
				sendText("w")
			case glfw.KeyE:
				sendText("e")
			case glfw.KeyR:
				sendText("r")
			case glfw.KeyT:
				sendText("t")
			case glfw.KeyY:
				sendText("y")
			case glfw.KeyU:
				sendText("u")
			case glfw.KeyI:
				sendText("i")
			case glfw.KeyO:
				sendText("o")
			case glfw.KeyP:
				sendText("p")
			case glfw.KeyA:
				sendText("a")
			case glfw.KeyS:
				sendText("s")
			case glfw.KeyD:
				sendText("d")
			case glfw.KeyF:
				sendText("f")
			case glfw.KeyG:
				sendText("g")
			case glfw.KeyH:
				sendText("h")
			case glfw.KeyJ:
				sendText("j")
			case glfw.KeyK:
				sendText("k")
			case glfw.KeyL:
				sendText("l")
			case glfw.KeyZ:
				sendText("z")
			case glfw.KeyX:
				sendText("x")
			case glfw.KeyC:
				sendText("c")
			case glfw.KeyV:
				sendText("v")
			case glfw.KeyB:
				sendText("b")
			case glfw.KeyN:
				sendText("n")
			case glfw.KeyM:
				sendText("m")
			case glfw.KeyEnter:
				sendText("enter")
			case glfw.KeyDelete:
				sendText("delete")
			case glfw.KeyBackspace:
				sendText("delete")
			case glfw.KeySpace:
				sendText(" ")
			}
		}
	}
}

func windowSizeCallback(w *glfw.Window, wd, ht int) {
	fwidth, fheight := scene.FramebufferSize(w)
	gl.Viewport(0, 0, int32(fwidth), int32(fheight))
}

func mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	player := universe.Player
	planet := player.Planet
	planetRen := universe.PlanetMap[planet.ID]
	if cursorGrabbed(w) {
		if action == glfw.Press && button == glfw.MouseButtonLeft {
			increment := player.LookDir().Mul(0.05)
			pos := player.Location()
			for i := 0; i < 100; i++ {
				pos = pos.Add(increment)
				cell := planet.CartesianToCell(pos)
				hitPlayer := false
				for _, otherPlayer := range universe.ConnectedPeople {
					if pos.Sub(otherPlayer.Position).Len() < 0.6 {
						log.Println(fmt.Sprintf("Hit %v", otherPlayer.Name))
						var ret bool
						universe.RPC.Go("API.HitPlayer", common.HitPlayerArgs{From: player.Name, Target: otherPlayer.Name, Amount: 1}, &ret, nil)
						hitPlayer = true
						break
					}
				}
				if hitPlayer {
					break
				}
				if cell != nil && cell.Material != common.Air {
					cellIndex := planet.CartesianToCellIndex(pos)
					planetRen.SetCellMaterial(cellIndex, common.Air, true)
					break
				}
			}
		} else if action == glfw.Press && button == glfw.MouseButtonRight {
			increment := player.LookDir().Mul(0.05)
			pos := player.Location()
			prevCellIndex := common.CellIndex{Lon: -1, Lat: -1, Alt: -1}
			cellIndex := common.CellIndex{}
			for i := 0; i < 100; i++ {
				pos = pos.Add(increment)
				nextCellIndex := planet.CartesianToCellIndex(pos)
				if nextCellIndex != cellIndex {
					prevCellIndex = cellIndex
					cellIndex = nextCellIndex
					cell := planet.CellIndexToCell(cellIndex)
					if cell != nil && cell.Material != common.Air {
						if prevCellIndex.Lon != -1 {
							planetRen.SetCellMaterial(prevCellIndex, player.Hotbar[player.ActiveHotBarSlot], true)
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
