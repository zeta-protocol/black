package simulation

// import (
// 	"fmt"
// 	"math/rand"

// 	"github.com/cosmos/cosmos-sdk/x/simulation"

// 	"github.com/zeta-protocol/black/x/issuance/types"
// )

// const (
// 	keyAssets = "Assets"
// )

// // ParamChanges defines the parameters that can be modified by param change proposals
// func ParamChanges(r *rand.Rand) []simulation.ParamChange {
// 	return []simulation.ParamChange{
// 		simulation.NewSimParamChange(types.ModuleName, keyAssets,
// 			func(r *rand.Rand) string {
// 				return fmt.Sprintf("\"%s\"", randomizedAssets(r))
// 			},
// 		),
// 	}
// }
