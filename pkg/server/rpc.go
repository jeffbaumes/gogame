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
