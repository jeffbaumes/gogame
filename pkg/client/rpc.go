package client

import "github.com/jeffbaumes/gogame/pkg/geom"

// API is the RPC tag for client calls
type API struct {
	planet *geom.Planet
	person *person
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
