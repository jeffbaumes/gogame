package client

import (
	"fmt"
	"log"
	"math"
	"time"

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
		guicusorposcallback := screen.CursorPosCallback()
		guicusorposcallback(w, xpos, ypos)
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

// func sendText(text string) {
// 	if text == "enter" {
// 		universe.Player.Intext = false
// 		// NewText()
// 		var ret bool
// 		universe.RPC.Go("API.SendText", universe.Player.Text, &ret, nil)
// 		universe.Player.Text = ""
// 		// need to send text to server
// 	} else if text == "delete" {
// 		if len(universe.Player.Text) != 0 {
// 			universe.Player.Text = universe.Player.Text[0:(len(universe.Player.Text) - 1)]
// 		}
// 	} else {
// 		universe.Player.Text = universe.Player.Text + text
// 	}
// }

var (
	oelapsed   = time.Duration(800000)
	ostart     = time.Now()
	guielapsed = time.Duration(800000)
	guistart   = time.Now()
)

func keyCallbackPlay(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	player := universe.Player
	switch action {
	case glfw.Press:
		switch key {
		case glfw.KeyEscape:
			player.Mode = "Options"
			w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		case op.InSky:
			player.Mode = "Apex"
		case op.OpText:
			player.Mode = "Text"
		case op.Inventory:
			player.Mode = "Inventory"
			w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		case op.Forwards:
			player.ForwardVel = player.WalkVel
		case op.Backwards:
			player.BackVel = player.WalkVel
		case op.Right:
			player.RightVel = player.WalkVel
		case op.Left:
			player.LeftVel = player.WalkVel
		case op.Mode:
			player.GameMode++
			if player.GameMode >= common.NumGameModes {
				player.GameMode = 0
			}
			if player.GameMode == common.Flying {
				player.FallVel = 0
			}
		case op.PlanetR:
			id := player.Planet.ID + 1
			if universe.PlanetMap[id] == nil {
				id = 0
			}
			player.Planet = universe.PlanetMap[id].Planet
			player.Spawn()
		case op.Up:
			if cursorGrabbed(w) {
				if player.GameMode == common.Normal {
					player.HoldingJump = true
				} else {
					player.UpVel = player.WalkVel
				}
			}
		case op.Down:
			if cursorGrabbed(w) && player.GameMode == common.Flying {
				player.DownVel = player.WalkVel
			}
		case op.PlanetL:
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
		}
	case glfw.Release:
		switch key {
		case op.Up:
			player.HoldingJump = false
			player.UpVel = 0
		case op.Down:
			player.DownVel = 0
		case op.Forwards:
			player.ForwardVel = 0
		case op.Backwards:
			player.BackVel = 0
		case op.Right:
			player.RightVel = 0
		case op.Left:
			player.LeftVel = 0
		}
	}
}

func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	player := universe.Player
	guikeycallback := screen.KeyCallBack()
	m := player.Mode
	if m == "Play" {
		keyCallbackPlay(w, key, scancode, action, mods)
	} else {
		if action == glfw.Press && key == glfw.KeyEscape {
			player.Mode = "Play"
			w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		} else if m == "Inventory" || m == "Text" || m == "Options" {
			guikeycallback(w, key, scancode, action, mods)
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

	guimousebuttoncallback := screen.MouseButtonCallback()
	planetRen := universe.PlanetMap[planet.ID]
	// keyCallbackPlay(w, glfw.Key(button), 0, action, mods)
	guimousebuttoncallback(w, button, action, mods)
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
		if action == glfw.Press && player.Mode != "Options" {
			player.Mode = "Play"
			w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		}
	}
}
