package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/zeta-protocol/black/x/earn/keeper"
	"github.com/zeta-protocol/black/x/earn/testutil"
	"github.com/zeta-protocol/black/x/earn/types"
	"github.com/stretchr/testify/suite"
)

type proposalTestSuite struct {
	testutil.Suite
}

func (suite *proposalTestSuite) SetupTest() {
	suite.Suite.SetupTest()
	suite.Keeper.SetParams(suite.Ctx, types.DefaultParams())
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(proposalTestSuite))
}

func (suite *proposalTestSuite) TestCommunityDepositProposal() {
	distKeeper := suite.App.GetDistrKeeper()
	ctx := suite.Ctx
	macc := distKeeper.GetDistributionAccount(ctx)
	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("ufury", 100000000))
	depositAmount := sdk.NewCoin("ufury", sdkmath.NewInt(10000000))
	suite.Require().NoError(suite.App.FundModuleAccount(ctx, macc.GetName(), fundAmount))
	feePool := distKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(fundAmount...)
	distKeeper.SetFeePool(ctx, feePool)
	suite.CreateVault("ufury", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)
	prop := types.NewCommunityPoolDepositProposal("test title",
		"desc", depositAmount)
	err := keeper.HandleCommunityPoolDepositProposal(ctx, suite.Keeper, prop)
	suite.Require().NoError(err)

	balance := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
	suite.Require().Equal(fundAmount.Sub(depositAmount), balance)
	feePool = distKeeper.GetFeePool(ctx)
	communityPoolBalance, change := feePool.CommunityPool.TruncateDecimal()
	suite.Require().Equal(fundAmount.Sub(depositAmount), communityPoolBalance)
	suite.Require().True(change.Empty())
}

func (suite *proposalTestSuite) TestCommunityWithdrawProposal() {
	distKeeper := suite.App.GetDistrKeeper()
	ctx := suite.Ctx
	macc := distKeeper.GetDistributionAccount(ctx)
	fundAmount := sdk.NewCoins(sdk.NewInt64Coin("ufury", 100000000))
	depositAmount := sdk.NewCoin("ufury", sdkmath.NewInt(10000000))
	suite.Require().NoError(suite.App.FundModuleAccount(ctx, macc.GetName(), fundAmount))
	feePool := distKeeper.GetFeePool(ctx)
	feePool.CommunityPool = sdk.NewDecCoinsFromCoins(fundAmount...)
	distKeeper.SetFeePool(ctx, feePool)
	// TODO update to STRATEGY_TYPE_SAVINGS once implemented
	suite.CreateVault("ufury", types.StrategyTypes{types.STRATEGY_TYPE_SAVINGS}, false, nil)
	deposit := types.NewCommunityPoolDepositProposal("test title",
		"desc", depositAmount)
	err := keeper.HandleCommunityPoolDepositProposal(ctx, suite.Keeper, deposit)
	suite.Require().NoError(err)

	balance := suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
	suite.Require().Equal(fundAmount.Sub(depositAmount), balance)

	withdraw := types.NewCommunityPoolWithdrawProposal("test title",
		"desc", depositAmount)
	err = keeper.HandleCommunityPoolWithdrawProposal(ctx, suite.Keeper, withdraw)
	suite.Require().NoError(err)
	balance = suite.BankKeeper.GetAllBalances(ctx, macc.GetAddress())
	suite.Require().Equal(fundAmount, balance)
	feePool = distKeeper.GetFeePool(ctx)
	communityPoolBalance, change := feePool.CommunityPool.TruncateDecimal()
	suite.Require().Equal(fundAmount, communityPoolBalance)
	suite.Require().True(change.Empty())
}
