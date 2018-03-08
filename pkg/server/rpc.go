package server

import (
	"log"

	"github.com/jeffbaumes/gogame/pkg/geom"
)

// API is the RPC tag for server calls
type API struct {
	connectedPeople []*connectedPerson
}

// GetChunk returns the planet chunk for the given chunk coordinates
func (api *API) GetChunk(args *geom.ChunkIndex, chunk *geom.Chunk) error {
	c := p.GetChunk(*args)
	if c != nil {
		*chunk = *c
	}
	return nil
}

// SetCellMaterial sets the material for a particular cell
func (api *API) SetCellMaterial(args *geom.RPCSetCellMaterialArgs, ret *bool) error {
	*ret = p.SetCellMaterial(args.Index, args.Material)
	var validPeople []*connectedPerson
	for _, c := range api.connectedPeople {
		var ret bool
		e := c.rpc.Call("API.SetCellMaterial", args, &ret)
		if e != nil {
			if e.Error() == "connection is shut down" {
				// Drop the client from the list
				// TODO: Should let the other clients know that this player is gone
				continue
			}
			log.Println("SetCellMaterial error:", e)
		}
		validPeople = append(validPeople, c)
	}
	api.connectedPeople = validPeople
	return nil
}
