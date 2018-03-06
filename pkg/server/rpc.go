package server

import (
	"log"

	"github.com/jeffbaumes/gogame/pkg/geom"
)

// Server is the RPC tag for server calls
type Server int

// GetChunk returns the planet chunk for the given chunk coordinates
func (t *Server) GetChunk(args *geom.ChunkIndex, chunk *geom.Chunk) error {
	c := p.GetChunk(*args)
	if c != nil {
		*chunk = *c
	}
	return nil
}

// SetCellMaterial sets the material for a particular cell
func (t *Server) SetCellMaterial(args *geom.RPCSetCellMaterialArgs, ret *bool) error {
	*ret = p.SetCellMaterial(args.Index, args.Material)
	for _, c := range clients {
		var ret bool
		e := c.Call("Client.SetCellMaterial", args, &ret)
		if e != nil {
			log.Println("SetCellMaterial error:", e)
		}
	}
	return nil
}
