package server

import "github.com/jeffbaumes/gogame/pkg/geom"

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

// SetCellMaterialArgs contains the arguments for the SetCellMaterial RPC call
type SetCellMaterialArgs struct {
	ind      geom.CellIndex
	material int
}

// SetCellMaterial sets the material for a particular cell
func (t *Server) SetCellMaterial(args *SetCellMaterialArgs, ret *bool) error {
	*ret = p.SetCellMaterial(args.ind, args.material)
	return nil
}
