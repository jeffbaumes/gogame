package client

import "github.com/jeffbaumes/gogame/pkg/geom"

// Client is the RPC tag for client calls
type Client int

// SetCellMaterial sets the material for a particular cell
func (t *Client) SetCellMaterial(args *geom.RPCSetCellMaterialArgs, ret *bool) error {
	*ret = p.SetCellMaterial(args.Index, args.Material)
	return nil
}
