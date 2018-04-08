package server

import (
	"errors"
	"log"

	"github.com/jeffbaumes/gogame/pkg/common"
)

// API is the RPC tag for server calls
type API struct {
	connectedPeople []*connectedPerson
}

// GetPlanetStates returns all planets
func (api *API) GetPlanetStates(args *int, states *[]*common.PlanetState) error {
	planets := []*common.PlanetState{}
	for _, planet := range universe.PlanetMap {
		planets = append(planets, &planet.PlanetState)
	}
	*states = planets
	return nil
}

// GetChunk returns the planet chunk for the given chunk coordinates
func (api *API) GetChunk(args *common.PlanetChunkIndex, chunk *common.Chunk) error {
	planet := universe.PlanetMap[args.Planet]
	if planet == nil {
		return errors.New("Unknown planet ID")
	}
	c := planet.GetChunk(args.ChunkIndex, false)
	if c != nil {
		*chunk = *c
	}
	return nil
}

// GetPlanetGeometry returns the low resolution geometry for a planet
func (api *API) GetPlanetGeometry(planetID *int, geom *common.PlanetGeometry) error {
	planet := universe.PlanetMap[*planetID]
	if planet == nil {
		return errors.New("Unknown planet ID")
	}
	g := planet.GetGeometry(false)
	if g != nil {
		*geom = *g
	}
	return nil
}

// UpdatePersonState updates a person's position
func (api *API) UpdatePersonState(state *common.PlayerState, ret *bool) error {
	var validPeople []*connectedPerson
	for _, c := range api.connectedPeople {
		if c.state.Name == state.Name {
			c.state = *state
		} else {
			var r bool
			e := c.rpc.Call("API.UpdatePersonState", state, &r)
			if e != nil {
				if e.Error() == "connection is shut down" {
					api.personDisconnected(c.state.Name)
					continue
				}
				log.Println("UpdatePersonState error:", e)
			}
		}
		validPeople = append(validPeople, c)
	}
	api.connectedPeople = validPeople
	*ret = true
	return nil
}

// SendText sends a text to all players
func (api *API) SendText(text *string, ret *bool) error {
	var validPeople []*connectedPerson
	for _, c := range api.connectedPeople {
		var r bool
		e := c.rpc.Call("API.SendText", text, &r)
		if e != nil {
			if e.Error() == "connection is shut down" {
				api.personDisconnected(c.state.Name)
				continue
			}
			log.Println("UpdatePersonState error:", e)
		}
		validPeople = append(validPeople, c)
	}
	api.connectedPeople = validPeople
	*ret = true
	return nil
}

// HitPlayer damages a person
func (api *API) HitPlayer(args *common.HitPlayerArgs, ret *bool) error {
	var validPeople []*connectedPerson
	for _, c := range api.connectedPeople {
		if c.state.Name == args.Target {
			var r bool
			e := c.rpc.Call("API.HitPlayer", args, &r)
			if e != nil {
				if e.Error() == "connection is shut down" {
					api.personDisconnected(c.state.Name)
					continue
				}
				log.Println("HitPlayer error:", e)
			}
		}
		validPeople = append(validPeople, c)
	}
	api.connectedPeople = validPeople
	*ret = true
	return nil
}

// SetCellMaterial sets the material for a particular cell
func (api *API) SetCellMaterial(args *common.RPCSetCellMaterialArgs, ret *bool) error {
	planet := universe.PlanetMap[args.Planet]
	if planet == nil {
		return errors.New("Unknown planet ID")
	}
	*ret = planet.SetCellMaterial(args.Index, args.Material, false)
	var validPeople []*connectedPerson
	for _, c := range api.connectedPeople {
		var ret bool
		e := c.rpc.Call("API.SetCellMaterial", args, &ret)
		if e != nil {
			if e.Error() == "connection is shut down" {
				api.personDisconnected(c.state.Name)
				continue
			}
			log.Println("SetCellMaterial error:", e)
		}
		validPeople = append(validPeople, c)
	}
	api.connectedPeople = validPeople
	return nil
}

func (api *API) personDisconnected(name string) {
	log.Printf("%v disconnected", name)
	for _, c := range api.connectedPeople {
		var ret bool
		c.rpc.Call("API.PersonDisconnected", name, &ret)
	}
}
