package client

import "github.com/jeffbaumes/gogame/pkg/geom"

// API is the RPC tag for client calls
type API struct {
	planet *geom.Planet
}

func newAPI(planet *geom.Planet) *API {
	return &API{planet}
}

// SetCellMaterial sets the material for a particular cell
func (api *API) SetCellMaterial(args *geom.RPCSetCellMaterialArgs, ret *bool) error {
	*ret = api.planet.SetCellMaterial(args.Index, args.Material)
	return nil
}
