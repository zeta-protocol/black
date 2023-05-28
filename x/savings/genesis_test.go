package savings_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/zeta-protocol/black/app"
	"github.com/zeta-protocol/black/x/savings"
	"github.com/zeta-protocol/black/x/savings/keeper"
	"github.com/zeta-protocol/black/x/savings/types"
)

type GenesisTestSuite struct {
	suite.Suite

	app     app.TestApp
	genTime time.Time
	ctx     sdk.Context
	keeper  keeper.Keeper
	addrs   []sdk.AccAddress
}

func (suite *GenesisTestSuite) SetupTest() {
	tApp := app.NewTestApp()
	suite.genTime = tmtime.Canonical(time.Date(2022, 1, 1, 1, 1, 1, 1, time.UTC))
	suite.ctx = tApp.NewContext(true, tmproto.Header{Height: 1, Time: suite.genTime})
	suite.keeper = tApp.GetSavingsKeeper()
	suite.app = tApp

	_, addrs := app.GeneratePrivKeyAddressPairs(3)
	suite.addrs = addrs
}

func (suite *GenesisTestSuite) TestInitExportGenesis() {
	params := types.NewParams(
		[]string{"btc", "ublack", "bnb"},
	)

	depositAmt := sdk.NewCoins(sdk.NewCoin("ublack", sdkmath.NewInt(1e8)))

	deposits := types.Deposits{
		types.NewDeposit(
			suite.addrs[0],
			depositAmt, // 100 ublack
		),
	}
	savingsGenesis := types.NewGenesisState(params, deposits)

	authBuilder := app.NewAuthBankGenesisBuilder().
		WithSimpleModuleAccount(types.ModuleAccountName, depositAmt)

	cdc := suite.app.AppCodec()
	suite.NotPanics(
		func() {
			suite.app.InitializeFromGenesisStatesWithTime(
				suite.genTime,
				authBuilder.BuildMarshalled(cdc),
				app.GenesisState{types.ModuleName: cdc.MustMarshalJSON(&savingsGenesis)},
			)
		},
	)

	expectedDeposits := suite.keeper.GetAllDeposits(suite.ctx)
	expectedGenesis := savingsGenesis
	expectedGenesis.Deposits = expectedDeposits
	exportedGenesis := savings.ExportGenesis(suite.ctx, suite.keeper)
	suite.Equal(expectedGenesis, exportedGenesis)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}