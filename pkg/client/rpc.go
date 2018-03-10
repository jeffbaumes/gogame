package client

import "github.com/jeffbaumes/gogame/pkg/geom"

// API is the RPC tag for client calls
type API struct {
	planet          *geom.Planet
	person          *person
	connectedPeople []*geom.PersonState
}

func newAPI(planet *geom.Planet, person *person) *API {
	return &API{planet: planet, person: person}
}

// SetCellMaterial sets the material for a particular cell
func (api *API) SetCellMaterial(args *geom.RPCSetCellMaterialArgs, ret *bool) error {
	*ret = api.planet.SetCellMaterial(args.Index, args.Material)
	return nil
}

// GetPersonState returns this client's logged in user state
func (api *API) GetPersonState(args *int, ret *geom.PersonState) error {
	ret.Name = api.person.name
	ret.Position = api.person.loc
	return nil
}

// PersonDisconnected notifies a client that a player has disconnected
func (api *API) PersonDisconnected(name *string, ret *bool) error {
	var validPeople []*geom.PersonState
	*ret = false
	for _, p := range api.connectedPeople {
		if p.Name != *name {
			validPeople = append(validPeople, p)
			*ret = true
		}
	}
	api.connectedPeople = validPeople
	return nil
}

// UpdatePersonState updates another person's state
func (api *API) UpdatePersonState(state *geom.PersonState, ret *bool) error {
	if state.Name == api.person.name {
		return nil
	}
	found := false
	for _, c := range api.connectedPeople {
		if c.Name == state.Name {
			c.Position = state.Position
			c.LookDir = state.LookDir
			found = true
			break
		}
	}
	if !found {
		api.connectedPeople = append(api.connectedPeople, state)
	}
	return nil
}
