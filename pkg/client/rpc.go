package client

import (
	"github.com/jeffbaumes/gogame/pkg/client/scene"
	"github.com/jeffbaumes/gogame/pkg/common"
)

// API is the RPC tag for client calls
type API struct {
	planetRen       *scene.Planet
	player          *common.Player
	connectedPeople []*common.PlayerState
}

func newAPI(planetRen *scene.Planet, player *common.Player) *API {
	return &API{planetRen: planetRen, player: player}
}

// SetCellMaterial sets the material for a particular cell
func (api *API) SetCellMaterial(args *common.RPCSetCellMaterialArgs, ret *bool) error {
	api.planetRen.SetCellMaterial(args.Index, args.Material)
	*ret = true
	return nil
}

// GetPersonState returns this client's logged in user state
func (api *API) GetPersonState(args *int, ret *common.PlayerState) error {
	ret.Name = api.player.Name
	ret.Position = api.player.Loc
	return nil
}

// PersonDisconnected notifies a client that a player has disconnected
func (api *API) PersonDisconnected(name *string, ret *bool) error {
	var validPeople []*common.PlayerState
	*ret = false
	for _, p := range api.connectedPeople {
		if p.Name != *name {
			validPeople = append(validPeople, p)
			*ret = true
		}
	}
	api.connectedPeople = validPeople
	return nil
}

// UpdatePersonState updates another person's state
func (api *API) UpdatePersonState(state *common.PlayerState, ret *bool) error {
	if state.Name == api.player.Name {
		return nil
	}
	found := false
	for _, c := range api.connectedPeople {
		if c.Name == state.Name {
			c.Position = state.Position
			c.LookDir = state.LookDir
			found = true
			break
		}
	}
	if !found {
		api.connectedPeople = append(api.connectedPeople, state)
	}
	return nil
}
