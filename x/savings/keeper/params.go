package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	liquidtypes "github.com/zeta-protocol/black/x/liquid/types"
	"github.com/zeta-protocol/black/x/savings/types"
)

const (
	bfuryDenom  = "bfury"
	bfuryPrefix = bfuryDenom + "-"
)

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramSubspace.GetParamSet(ctx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// IsDenomSupported returns a boolean indicating if a denom is supported
func (k Keeper) IsDenomSupported(ctx sdk.Context, denom string) bool {
	p := k.GetParams(ctx)
	for _, supportedDenom := range p.SupportedDenoms {
		if supportedDenom == denom {
			return true
		}

		if supportedDenom == liquidtypes.DefaultDerivativeDenom {
			if k.liquidKeeper.IsDerivativeDenom(ctx, denom) {
				return true
			}
		}
	}

	return false
}
