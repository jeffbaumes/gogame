package server

import (
	"errors"
)

type GetChunkArgs struct {
	Lat, Lon, Alt int
}

type Server int

func (t *Server) GetChunk(args *GetChunkArgs, chunk *Chunk) error {
	c := p.GetChunk(args.Lat, args.Lon, args.Alt)
	if c != nil {
		*chunk = *c
	}
	return nil
}

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Arith) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}
