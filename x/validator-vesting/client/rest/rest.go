package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

// RegisterRoutes registers blackdist-related REST handlers to a router
func RegisterRoutes(cliCtx client.Context, rtr *mux.Router) {
	registerQueryRoutes(cliCtx, rtr)
}
