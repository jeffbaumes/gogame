package common

import (
	opensimplex "github.com/ojrac/opensimplex-go"
)

// Universe stores the set of planets in a universe
type Universe struct {
	seed         int
	nextPlanetID int
	noise        *opensimplex.Noise
	PlanetMap    map[int]*Planet
}

// NewUniverse creates a universe with a given seed
func NewUniverse(seed int) *Universe {
	u := Universe{}
	u.noise = opensimplex.NewWithSeed(0)
	u.PlanetMap = make(map[int]*Planet)
	return &u
}

// AddPlanet adds a planet to the universe and sets its ID
func (u *Universe) AddPlanet(planet *Planet) {
	planet.ID = u.nextPlanetID
	u.PlanetMap[planet.ID] = planet
	u.nextPlanetID++
}
