package scene

import (
	"math"
	"net/rpc"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jeffbaumes/buildorb/pkg/common"
)

// Universe stores the state of the universe
type Universe struct {
	Player          *common.Player
	PlanetMap       map[int]*Planet
	ConnectedPeople []*common.PlayerState
	RPC             *rpc.Client
}

// NewUniverse creates a new universe
func NewUniverse(player *common.Player, rpc *rpc.Client) *Universe {
	u := Universe{}
	u.Player = player
	u.PlanetMap = make(map[int]*Planet)
	u.RPC = rpc
	return &u
}

// AddPlanet adds a planet to the planet map
func (u *Universe) AddPlanet(planet *Planet) {
	u.PlanetMap[planet.Planet.ID] = planet
}

// Draw draws the universe's planets
func (u *Universe) Draw(w *glfw.Window, time float64) {
	player := u.Player
	loc := player.Location()
	planetRen := u.PlanetMap[player.Planet.ID]
	planetRotation := time / planetRen.Planet.RotationSeconds
	planetRotation *= 2 * math.Pi
	orbitPosition := time / planetRen.Planet.OrbitSeconds
	orbitPosition *= 2 * math.Pi
	planetLoc := planetRen.location(time, u.PlanetMap)
	planetRotateNeg := mgl32.Rotate3DZ(-float32(planetRotation))

	rotated := planetRotateNeg.Mul3x1(loc)
	sunDir := planetLoc.Add(rotated).Normalize()

	vpnDotSun := float64(rotated.Normalize().Dot(sunDir))
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
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	for _, planetRen := range u.PlanetMap {
		planetRen.Draw(u.Player, u.PlanetMap, w, time)
	}
}
