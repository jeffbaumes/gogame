package common

import (
	"bytes"
	"database/sql"
	"encoding/gob"

	opensimplex "github.com/ojrac/opensimplex-go"
)

// Universe stores the set of planets in a universe
type Universe struct {
	seed      int
	noise     *opensimplex.Noise
	PlanetMap map[int]*Planet
}

// NewUniverse creates a universe with a given seed
func NewUniverse(db *sql.DB, systemType string) *Universe {
	u := Universe{}
	u.noise = opensimplex.NewWithSeed(0)
	u.PlanetMap = make(map[int]*Planet)
	planetStates := queryPlanetStates(db)

	// If no planets in the database, generate a planetary system
	if len(planetStates) == 0 {
		initializeGenerators()
		systemGen := systems[systemType]
		if systemGen == nil {
			systemGen = systems["planet"]
		}
		planetStates = systemGen()
		for _, state := range planetStates {
			savePlanetState(db, *state)
		}
	}

	// Put the planets in the universe
	for _, state := range planetStates {
		planet := NewPlanet(*state, nil, db)
		u.PlanetMap[planet.ID] = planet
	}

	return &u
}

func queryPlanetStates(db *sql.DB) []*PlanetState {
	states := []*PlanetState{}
	rows, err := db.Query("SELECT data FROM planet")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var val PlanetState
		var data []byte
		err = rows.Scan(&data)
		if err != nil {
			panic(err)
		}
		var dbuf bytes.Buffer
		dbuf.Write(data)
		dec := gob.NewDecoder(&dbuf)
		err = dec.Decode(&val)
		if err != nil {
			panic(err)
		}
		states = append(states, &val)
	}
	return states
}

func savePlanetState(db *sql.DB, state PlanetState) {
	stmt, err := db.Prepare("INSERT INTO planet VALUES (?, ?)")
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(state)
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec(state.ID, buf.Bytes())
	if err != nil {
		panic(err)
	}
}
