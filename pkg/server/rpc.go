package server

// Server is the RPC tag for server calls
type Server int

// GetChunkArgs holds the arguments for a call to GetChunk
type GetChunkArgs struct {
	Lon, Lat, Alt int
}

// GetChunk returns the planet chunk for the given chunk coordinates
func (t *Server) GetChunk(args *GetChunkArgs, chunk *Chunk) error {
	c := p.GetChunk(args.Lon, args.Lat, args.Alt)
	if c != nil {
		*chunk = *c
	}
	return nil
}
