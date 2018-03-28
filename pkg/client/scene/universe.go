package scene

import (
	"net/rpc"

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
