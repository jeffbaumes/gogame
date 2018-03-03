package server

import (
	opensimplex "github.com/ojrac/opensimplex-go"
)

type Universe struct {
	seed      int
	noise     *opensimplex.Noise
	planetMap map[int]*Planet
}

func NewUniverse(seed int) *Universe {
	u := Universe{}
	u.noise = opensimplex.NewWithSeed(0)
	return &u
}
