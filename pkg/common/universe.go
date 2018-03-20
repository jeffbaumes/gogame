package common

import (
	opensimplex "github.com/ojrac/opensimplex-go"
)

// Universe stores the set of planets in a universe
type Universe struct {
	seed      int
	noise     *opensimplex.Noise
	planetMap map[int]*Planet
}

// NewUniverse creates a universe with a given seed
func NewUniverse(seed int) *Universe {
	u := Universe{}
	u.noise = opensimplex.NewWithSeed(0)
	return &u
}
