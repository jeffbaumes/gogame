package scene

import (
	"net/rpc"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jeffbaumes/gogame/pkg/common"
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
	for _, planetRen := range u.PlanetMap {
		planetRen.Draw(u.Player, u.PlanetMap, w, time)
	}
}
