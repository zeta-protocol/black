package ante_test

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/zeta-protocol/black/app"
	"github.com/zeta-protocol/black/app/ante"
)

func mustParseDecCoins(value string) sdk.DecCoins {
	coins, err := sdk.ParseDecCoins(strings.ReplaceAll(value, ";", ","))
	if err != nil {
		panic(err)
	}

	return coins
}

func TestEvmMinGasFilter(t *testing.T) {
	tApp := app.NewTestApp()
	handler := ante.NewEvmMinGasFilter(tApp.GetEvmKeeper())

	ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})
	tApp.GetEvmKeeper().SetParams(ctx, evmtypes.Params{
		EvmDenom: "ablack",
	})

	testCases := []struct {
		name                 string
		minGasPrices         sdk.DecCoins
		expectedMinGasPrices sdk.DecCoins
	}{
		{
			"no min gas prices",
			mustParseDecCoins(""),
			mustParseDecCoins(""),
		},
		{
			"zero ublack gas price",
			mustParseDecCoins("0ublack"),
			mustParseDecCoins("0ublack"),
		},
		{
			"non-zero ublack gas price",
			mustParseDecCoins("0.001ublack"),
			mustParseDecCoins("0.001ublack"),
		},
		{
			"zero ublack gas price, min ablack price",
			mustParseDecCoins("0ublack;100000ablack"),
			mustParseDecCoins("0ublack"), // ablack is removed
		},
		{
			"zero ublack gas price, min ablack price, other token",
			mustParseDecCoins("0ublack;100000ablack;0.001other"),
			mustParseDecCoins("0ublack;0.001other"), // ablack is removed
		},
		{
			"non-zero ublack gas price, min ablack price",
			mustParseDecCoins("0.25ublack;100000ablack;0.001other"),
			mustParseDecCoins("0.25ublack;0.001other"), // ablack is removed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tApp.NewContext(true, tmproto.Header{Height: 1, Time: tmtime.Now()})

			ctx = ctx.WithMinGasPrices(tc.minGasPrices)
			mmd := MockAnteHandler{}

			_, err := handler.AnteHandle(ctx, nil, false, mmd.AnteHandle)
			require.NoError(t, err)
			require.True(t, mmd.WasCalled)

			assert.NoError(t, mmd.CalledCtx.MinGasPrices().Validate())
			assert.Equal(t, tc.expectedMinGasPrices, mmd.CalledCtx.MinGasPrices())
		})
	}
}
