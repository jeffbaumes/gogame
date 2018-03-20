package server

import (
	"log"

	"github.com/jeffbaumes/gogame/pkg/common"
)

// API is the RPC tag for server calls
type API struct {
	connectedPeople []*connectedPerson
}

// GetChunk returns the planet chunk for the given chunk coordinates
func (api *API) GetChunk(args *common.ChunkIndex, chunk *common.Chunk) error {
	c := p.GetChunk(*args)
	if c != nil {
		*chunk = *c
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
	return nil
}

// SetCellMaterial sets the material for a particular cell
func (api *API) SetCellMaterial(args *common.RPCSetCellMaterialArgs, ret *bool) error {
	*ret = p.SetCellMaterial(args.Index, args.Material)
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
