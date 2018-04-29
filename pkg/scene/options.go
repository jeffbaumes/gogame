package scene

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/anbcodes/goguigl/gui"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jeffbaumes/gogame/pkg/common"
)

// Options holds all the user-defined options for controlling the game
type Options struct {
	Walk, Run, Left, Right, Forwards, Backwards, FlySprint, Up, Down, Inventory, Slot1, Slot2, Slot3, Slot4, Slot5, Slot6, Slot7, Slot8, Slot9, Slot10, Slot11, Slot12, Mode, OpText, InSky, PlanetR, PlanetL, Destroy, Build glfw.Key
	walke, runee, lefte, righte, forwardse, backwardse, flysprinte, upe, downe, inventorye, builde, destroye                                                                                                                  *gui.Entry
	walkl, runl, leftl, rightl, forwardsl, backwardsl, flysprintl, upl, downl, inventoryl, buildl, destroyl                                                                                                                   *gui.Label
	Everythinge                                                                                                                                                                                                               []*gui.Entry
	Everythingl                                                                                                                                                                                                               []*gui.Label
}

func writefile(t string) {
	err := ioutil.WriteFile("options.txt", []byte(t), 0644)
	if err != nil {
		panic(err)
	}
}

func readfile() string {
	b, err := ioutil.ReadFile("options.txt")
	if err != nil {
		os.Create("options.txt")
		return readfile()
	}
	return string(b)
}

func load(o *Options) {
	f2 := strings.Split(readfile(), "\n")[0]
	s := strings.Split(f2, ";")
	for _, h := range s {
		g := strings.Split(h, "=")
		if g[1] == "" {
			continue
		}
		fmt.Printf("'%s'\n", g[1])
		n, err := strconv.Atoi(g[1])
		if err != nil {
			panic(err)
		}
		f := glfw.Key(n)
		switch g[0] {
		case "forward":
			o.Forwards = f
		case "backward":
			o.Backwards = f
		case "left":
			o.Left = f
		case "right":
			o.Right = f
		case "run":
			o.Run = f
		case "flyrun":
			o.FlySprint = f
		case "up":
			o.Up = f
		case "down":
			o.Down = f
		case "inventory":
			o.Inventory = f
		case "build":
			o.Build = f
		case "destroy":
			o.Destroy = f
		}
	}
}

// NewOptions creates a new options GUI
func NewOptions(screen *gui.Screen) *Options {
	o := Options{}
	o.InSky = glfw.KeyK
	o.Mode = glfw.KeyM
	o.PlanetL = glfw.KeyLeftBracket
	o.PlanetR = glfw.KeyRightBracket
	o.OpText = glfw.KeyEnter
	o.Slot1 = glfw.Key1
	o.Slot2 = glfw.Key2
	o.Slot3 = glfw.Key3
	o.Slot4 = glfw.Key4
	o.Slot5 = glfw.Key5
	o.Slot6 = glfw.Key6
	o.Slot7 = glfw.Key7
	o.Slot8 = glfw.Key8
	o.Slot9 = glfw.Key9
	o.Slot10 = glfw.Key0
	o.Slot11 = glfw.KeyMinus
	o.Slot12 = glfw.KeyEqual
	o.Forwards = glfw.KeyW
	o.Backwards = glfw.KeyS
	o.Left = glfw.KeyA
	o.Right = glfw.KeyD
	o.Up = glfw.KeySpace
	o.Down = glfw.KeyLeftShift
	o.Run = glfw.KeyLeftShift
	o.FlySprint = glfw.KeyLeftControl
	o.Inventory = glfw.KeyE
	o.Build = glfw.Key(1)
	o.Destroy = glfw.Key(0)
	load(&o)
	o.forwardse = gui.NewKeyEntry(screen, o.Forwards, -0.50, 0.75, 0.2, 0.15, 0.05, nil)
	o.forwardsl = gui.NewLabel(screen, "Forward:", -0.95, 0.80, 0.09)
	o.backwardse = gui.NewKeyEntry(screen, o.Backwards, -0.50, 0.55, 0.2, 0.15, 0.05, nil)
	o.backwardsl = gui.NewLabel(screen, "Backwards:", -0.95, 0.60, 0.09)
	o.lefte = gui.NewKeyEntry(screen, o.Left, -0.50, 0.35, 0.2, 0.15, 0.05, nil)
	o.leftl = gui.NewLabel(screen, "Left:", -0.95, 0.40, 0.09)
	o.righte = gui.NewKeyEntry(screen, o.Right, -0.50, 0.15, 0.2, 0.15, 0.05, nil)
	o.rightl = gui.NewLabel(screen, "Right:", -0.95, 0.20, 0.09)
	o.upe = gui.NewKeyEntry(screen, o.Up, 0, 0.75, 0.2, 0.15, 0.05, nil)
	o.upl = gui.NewLabel(screen, "Up:", -0.25, 0.80, 0.09)
	o.downe = gui.NewKeyEntry(screen, o.Down, 0, 0.55, 0.2, 0.15, 0.05, nil)
	o.downl = gui.NewLabel(screen, "Down:", -0.25, 0.60, 0.09)
	o.runee = gui.NewKeyEntry(screen, o.Run, 0.75, 0.75, 0.2, 0.15, 0.05, nil)
	o.runl = gui.NewLabel(screen, "Run:", 0.25, 0.80, 0.09)
	o.flysprinte = gui.NewKeyEntry(screen, o.FlySprint, 0.75, 0.55, 0.2, 0.15, 0.05, nil)
	o.flysprintl = gui.NewLabel(screen, "FlyRun:", 0.25, 0.60, 0.09)
	o.inventorye = gui.NewKeyEntry(screen, o.Inventory, 0.75, 0.35, 0.2, 0.15, 0.05, nil)
	o.inventoryl = gui.NewLabel(screen, "Inventory:", 0.25, 0.40, 0.09)
	o.builde = gui.NewKeyEntry(screen, o.Build, 0.75, -0.05, 0.2, 0.15, 0.05, nil)
	o.buildl = gui.NewLabel(screen, "Build:", 0.25, 0, 0.09)
	o.destroye = gui.NewKeyEntry(screen, o.Destroy, 0.75, 0.15, 0.2, 0.15, 0.05, nil)
	o.destroyl = gui.NewLabel(screen, "Destroy:", 0.25, 0.20, 0.09)
	o.Everythinge = append(o.Everythinge, o.forwardse, o.backwardse, o.righte, o.lefte, o.upe, o.downe, o.runee, o.flysprinte, o.inventorye, o.builde, o.destroye)
	o.Everythingl = append(o.Everythingl, o.forwardsl, o.backwardsl, o.rightl, o.leftl, o.upl, o.downl, o.runl, o.flysprintl, o.inventoryl, o.buildl, o.destroyl)
	fmt.Printf("%+v", o)
	return &o
}

var oldop string

// Draw draws the options screen
func (o *Options) Draw(p *common.Player) {
	if oldop != p.Mode {
		oldop = p.Mode
		o.Forwards = o.forwardse.Key
		o.Backwards = o.backwardse.Key
		o.Left = o.lefte.Key
		o.Right = o.righte.Key
		o.Inventory = o.inventorye.Key
		o.Run = o.runee.Key
		o.FlySprint = o.flysprinte.Key
		o.Up = o.upe.Key
		o.Down = o.downe.Key
		o.Build = o.builde.Key
		o.Destroy = o.destroye.Key
	}
	if p.Mode == "Options" {
		for _, x := range o.Everythinge {
			x.Hide = false
		}
		for _, x := range o.Everythingl {
			x.Hide = false
		}
	} else {
		for _, x := range o.Everythinge {
			x.Hide = true
		}
		for _, x := range o.Everythingl {
			x.Hide = true
		}
	}
}
