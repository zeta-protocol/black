package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/zeta-protocol/black/app"
	earntypes "github.com/zeta-protocol/black/x/earn/types"
	"github.com/zeta-protocol/black/x/incentive/testutil"
	"github.com/zeta-protocol/black/x/incentive/types"
)

type AccumulateEarnRewardsIntegrationTests struct {
	testutil.IntegrationTester

	keeper    TestKeeper
	userAddrs []sdk.AccAddress
	valAddrs  []sdk.ValAddress
}

func TestAccumulateEarnRewardsIntegrationTests(t *testing.T) {
	suite.Run(t, new(AccumulateEarnRewardsIntegrationTests))
}

func (suite *AccumulateEarnRewardsIntegrationTests) SetupTest() {
	suite.IntegrationTester.SetupTest()

	suite.keeper = TestKeeper{
		Keeper: suite.App.GetIncentiveKeeper(),
	}

	_, addrs := app.GeneratePrivKeyAddressPairs(5)
	suite.userAddrs = addrs[0:2]
	suite.valAddrs = []sdk.ValAddress{
		sdk.ValAddress(addrs[2]),
		sdk.ValAddress(addrs[3]),
	}

	// Setup app with test state
	authBuilder := app.NewAuthBankGenesisBuilder().
		WithSimpleAccount(addrs[0], cs(c("ublack", 1e12))).
		WithSimpleAccount(addrs[1], cs(c("ublack", 1e12))).
		WithSimpleAccount(addrs[2], cs(c("ublack", 1e12))).
		WithSimpleAccount(addrs[3], cs(c("ublack", 1e12)))

	incentiveBuilder := testutil.NewIncentiveGenesisBuilder().
		WithGenesisTime(suite.GenesisTime).
		WithSimpleEarnRewardPeriod("bblack", cs())

	savingsBuilder := testutil.NewSavingsGenesisBuilder().
		WithSupportedDenoms("bblack")

	earnBuilder := testutil.NewEarnGenesisBuilder().
		WithAllowedVaults(earntypes.AllowedVault{
			Denom:             "bblack",
			Strategies:        earntypes.StrategyTypes{earntypes.STRATEGY_TYPE_SAVINGS},
			IsPrivateVault:    false,
			AllowedDepositors: nil,
		})

	stakingBuilder := testutil.NewStakingGenesisBuilder()

	mintBuilder := testutil.NewMintGenesisBuilder().
		WithInflationMax(sdk.OneDec()).
		WithInflationMin(sdk.OneDec()).
		WithMinter(sdk.OneDec(), sdk.ZeroDec()).
		WithMintDenom("ublack")

	suite.StartChainWithBuilders(
		authBuilder,
		incentiveBuilder,
		savingsBuilder,
		earnBuilder,
		stakingBuilder,
		mintBuilder,
	)
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestStateUpdatedWhenBlockTimeHasIncreased() {
	suite.AddIncentiveEarnMultiRewardPeriod(
		types.NewMultiRewardPeriod(
			true,
			"bblack",         // reward period is set for "bblack" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ublack", 1000)), // same denoms as in global indexes
		),
	)

	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ublack", 800000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ublack", 200000))
	suite.NoError(err)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], derivative0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[1], derivative1, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: derivative0.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ublack",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: derivative1.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ublack",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}

	suite.keeper.storeGlobalEarnIndexes(suite.Ctx, globalIndexes)
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative1.Denom, suite.Ctx.BlockTime())

	val0 := suite.GetAbciValidator(suite.valAddrs[0])
	val1 := suite.GetAbciValidator(suite.valAddrs[1])

	// Mint tokens, distribute to validators, claim staking rewards
	// 1 hour later
	_, resBeginBlock := suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)

	validatorRewards, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	suite.Require().Contains(validatorRewards, suite.valAddrs[1].String(), "there should be claim events for validator 0")
	suite.Require().Contains(validatorRewards, suite.valAddrs[0].String(), "there should be claim events for validator 1")

	// check time and factors

	suite.StoredEarnTimeEquals(derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(derivative1.Denom, suite.Ctx.BlockTime())

	stakingRewardIndexes0 := sdk.NewDecFromInt(validatorRewards[suite.valAddrs[0].String()].
		AmountOf("ublack")).
		Quo(sdk.NewDecFromInt(derivative0.Amount))

	stakingRewardIndexes1 := sdk.NewDecFromInt(validatorRewards[suite.valAddrs[1].String()].
		AmountOf("ublack")).
		Quo(sdk.NewDecFromInt(derivative1.Amount))

	suite.StoredEarnIndexesEqual(derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ublack",
			RewardFactor:   d("3.64").Add(stakingRewardIndexes0),
		},
	})
	suite.StoredEarnIndexesEqual(derivative1.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("7.22"),
		},
		{
			CollateralType: "ublack",
			RewardFactor:   d("3.64").Add(stakingRewardIndexes1),
		},
	})
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestStateUpdatedWhenBlockTimeHasIncreased_partialDeposit() {
	suite.AddIncentiveEarnMultiRewardPeriod(
		types.NewMultiRewardPeriod(
			true,
			"bblack",         // reward period is set for "bblack" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ublack", 1000)), // same denoms as in global indexes
		),
	)

	// 800000bblack0 minted, 700000 deposited
	// 200000bblack1 minted, 100000 deposited
	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ublack", 800000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ublack", 200000))
	suite.NoError(err)

	depositAmount0 := c(derivative0.Denom, 700000)
	depositAmount1 := c(derivative1.Denom, 100000)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], depositAmount0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[1], depositAmount1, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	globalIndexes := types.MultiRewardIndexes{
		{
			CollateralType: derivative0.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ublack",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: derivative1.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ublack",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}

	suite.keeper.storeGlobalEarnIndexes(suite.Ctx, globalIndexes)

	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative1.Denom, suite.Ctx.BlockTime())

	val0 := suite.GetAbciValidator(suite.valAddrs[0])
	val1 := suite.GetAbciValidator(suite.valAddrs[1])

	// Mint tokens, distribute to validators, claim staking rewards
	// 1 hour later
	_, resBeginBlock := suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)

	validatorRewards, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	suite.Require().Contains(validatorRewards, suite.valAddrs[1].String(), "there should be claim events for validator 0")
	suite.Require().Contains(validatorRewards, suite.valAddrs[0].String(), "there should be claim events for validator 1")

	// check time and factors

	suite.StoredEarnTimeEquals(derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(derivative1.Denom, suite.Ctx.BlockTime())

	// Divided by deposit amounts, not bank supply amounts
	stakingRewardIndexes0 := sdk.NewDecFromInt(validatorRewards[suite.valAddrs[0].String()].
		AmountOf("ublack")).
		Quo(sdk.NewDecFromInt(depositAmount0.Amount))

	stakingRewardIndexes1 := sdk.NewDecFromInt(validatorRewards[suite.valAddrs[1].String()].
		AmountOf("ublack")).
		Quo(sdk.NewDecFromInt(depositAmount1.Amount))

	// Slightly increased rewards due to less bblack deposited
	suite.StoredEarnIndexesEqual(derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("8.248571428571428571"),
		},
		{
			CollateralType: "ublack",
			RewardFactor:   d("4.154285714285714285").Add(stakingRewardIndexes0),
		},
	})

	suite.StoredEarnIndexesEqual(derivative1.Denom, types.RewardIndexes{
		{
			CollateralType: "earn",
			RewardFactor:   d("14.42"),
		},
		{
			CollateralType: "ublack",
			RewardFactor:   d("7.24").Add(stakingRewardIndexes1),
		},
	})
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestStateUnchangedWhenBlockTimeHasNotIncreased() {
	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ublack", 1000000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ublack", 1000000))
	suite.NoError(err)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], derivative0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[1], derivative1, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: derivative0.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ublack",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: derivative1.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ublack",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.keeper.storeGlobalEarnIndexes(suite.Ctx, previousIndexes)

	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative1.Denom, suite.Ctx.BlockTime())

	period := types.NewMultiRewardPeriod(
		true,
		"bblack",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(c("earn", 2000), c("ublack", 1000)), // same denoms as in global indexes
	)

	// Must manually accumulate rewards as BeginBlockers only run when the block time increases
	// This does not run any x/mint or x/distribution BeginBlockers
	suite.keeper.AccumulateEarnRewards(suite.Ctx, period)

	// check time and factors

	suite.StoredEarnTimeEquals(derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(derivative1.Denom, suite.Ctx.BlockTime())

	expected, f := previousIndexes.Get(derivative0.Denom)
	suite.True(f)
	suite.StoredEarnIndexesEqual(derivative0.Denom, expected)

	expected, f = previousIndexes.Get(derivative1.Denom)
	suite.True(f)
	suite.StoredEarnIndexesEqual(derivative1.Denom, expected)
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestNoAccumulationWhenSourceSharesAreZero() {
	suite.AddIncentiveEarnMultiRewardPeriod(
		types.NewMultiRewardPeriod(
			true,
			"bblack",         // reward period is set for "bblack" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ublack", 1000)), // same denoms as in global indexes
		),
	)

	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ublack", 1000000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ublack", 1000000))
	suite.NoError(err)

	// No earn deposits

	previousIndexes := types.MultiRewardIndexes{
		{
			CollateralType: derivative0.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ublack",
					RewardFactor:   d("0.04"),
				},
			},
		},
		{
			CollateralType: derivative1.Denom,
			RewardIndexes: types.RewardIndexes{
				{
					CollateralType: "earn",
					RewardFactor:   d("0.02"),
				},
				{
					CollateralType: "ublack",
					RewardFactor:   d("0.04"),
				},
			},
		},
	}
	suite.keeper.storeGlobalEarnIndexes(suite.Ctx, previousIndexes)

	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative0.Denom, suite.Ctx.BlockTime())
	suite.keeper.SetEarnRewardAccrualTime(suite.Ctx, derivative1.Denom, suite.Ctx.BlockTime())

	val0 := suite.GetAbciValidator(suite.valAddrs[0])
	val1 := suite.GetAbciValidator(suite.valAddrs[1])

	// Mint tokens, distribute to validators, claim staking rewards
	// 1 hour later
	_, _ = suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)
	// check time and factors

	suite.StoredEarnTimeEquals(derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(derivative1.Denom, suite.Ctx.BlockTime())

	expected, f := previousIndexes.Get(derivative0.Denom)
	suite.True(f)
	suite.StoredEarnIndexesEqual(derivative0.Denom, expected)

	expected, f = previousIndexes.Get(derivative1.Denom)
	suite.True(f)
	suite.StoredEarnIndexesEqual(derivative1.Denom, expected)
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestStateAddedWhenStateDoesNotExist() {
	suite.AddIncentiveEarnMultiRewardPeriod(
		types.NewMultiRewardPeriod(
			true,
			"bblack",         // reward period is set for "bblack" to apply to all vaults
			time.Unix(0, 0), // ensure the test is within start and end times
			distantFuture,
			cs(c("earn", 2000), c("ublack", 1000)), // same denoms as in global indexes
		),
	)

	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ublack", 1000000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ublack", 1000000))
	suite.NoError(err)

	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[0], derivative0, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)
	err = suite.DeliverEarnMsgDeposit(suite.userAddrs[1], derivative1, earntypes.STRATEGY_TYPE_SAVINGS)
	suite.NoError(err)

	val0 := suite.GetAbciValidator(suite.valAddrs[0])
	val1 := suite.GetAbciValidator(suite.valAddrs[1])

	_, resBeginBlock := suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)

	// After the second accumulation both current block time and indexes should be stored.
	suite.StoredEarnTimeEquals(derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(derivative1.Denom, suite.Ctx.BlockTime())

	validatorRewards0, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	firstStakingRewardIndexes0 := sdk.NewDecFromInt(validatorRewards0[suite.valAddrs[0].String()].
		AmountOf("ublack")).
		Quo(sdk.NewDecFromInt(derivative0.Amount))

	firstStakingRewardIndexes1 := sdk.NewDecFromInt(validatorRewards0[suite.valAddrs[1].String()].
		AmountOf("ublack")).
		Quo(sdk.NewDecFromInt(derivative1.Amount))

	// After the first accumulation only the current block time should be stored.
	// The indexes will be empty as no time has passed since the previous block because it didn't exist.
	suite.StoredEarnTimeEquals(derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(derivative1.Denom, suite.Ctx.BlockTime())

	// First accumulation can have staking rewards, but no other rewards
	suite.StoredEarnIndexesEqual(derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "ublack",
			RewardFactor:   firstStakingRewardIndexes0,
		},
	})
	suite.StoredEarnIndexesEqual(derivative1.Denom, types.RewardIndexes{
		{
			CollateralType: "ublack",
			RewardFactor:   firstStakingRewardIndexes1,
		},
	})

	_, resBeginBlock = suite.NextBlockAfterWithReq(
		1*time.Hour,
		abci.RequestEndBlock{},
		abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{
					{
						Validator:       val0,
						SignedLastBlock: true,
					},
					{
						Validator:       val1,
						SignedLastBlock: true,
					},
				},
			},
		},
	)

	// After the second accumulation both current block time and indexes should be stored.
	suite.StoredEarnTimeEquals(derivative0.Denom, suite.Ctx.BlockTime())
	suite.StoredEarnTimeEquals(derivative1.Denom, suite.Ctx.BlockTime())

	validatorRewards1, _ := suite.GetBeginBlockClaimedStakingRewards(resBeginBlock)

	secondStakingRewardIndexes0 := sdk.NewDecFromInt(validatorRewards1[suite.valAddrs[0].String()].
		AmountOf("ublack")).
		Quo(sdk.NewDecFromInt(derivative0.Amount))

	secondStakingRewardIndexes1 := sdk.NewDecFromInt(validatorRewards1[suite.valAddrs[1].String()].
		AmountOf("ublack")).
		Quo(sdk.NewDecFromInt(derivative1.Amount))

	// Second accumulation has both staking rewards and incentive rewards
	// ublack incentive rewards: 3600 * 1000 / (2 * 1000000) == 1.8
	suite.StoredEarnIndexesEqual(derivative0.Denom, types.RewardIndexes{
		{
			CollateralType: "ublack",
			// Incentive rewards + both staking rewards
			RewardFactor: d("1.8").Add(firstStakingRewardIndexes0).Add(secondStakingRewardIndexes0),
		},
		{
			CollateralType: "earn",
			RewardFactor:   d("3.6"),
		},
	})
	suite.StoredEarnIndexesEqual(derivative1.Denom, types.RewardIndexes{
		{
			CollateralType: "ublack",
			// Incentive rewards + both staking rewards
			RewardFactor: d("1.8").Add(firstStakingRewardIndexes1).Add(secondStakingRewardIndexes1),
		},
		{
			CollateralType: "earn",
			RewardFactor:   d("3.6"),
		},
	})
}

func (suite *AccumulateEarnRewardsIntegrationTests) TestNoPanicWhenStateDoesNotExist() {
	derivative0, err := suite.MintLiquidAnyValAddr(suite.userAddrs[0], suite.valAddrs[0], c("ublack", 1000000))
	suite.NoError(err)
	derivative1, err := suite.MintLiquidAnyValAddr(suite.userAddrs[1], suite.valAddrs[1], c("ublack", 1000000))
	suite.NoError(err)

	period := types.NewMultiRewardPeriod(
		true,
		"bblack",
		time.Unix(0, 0), // ensure the test is within start and end times
		distantFuture,
		cs(),
	)

	// Accumulate with no earn shares and no rewards per second will result in no increment to the indexes.
	// No increment and no previous indexes stored, results in an updated of nil. Setting this in the state panics.
	// Check there is no panic.
	suite.NotPanics(func() {
		// This does not update any state, as there are no bblack vaults
		// to iterate over, denoms are unknown
		suite.keeper.AccumulateEarnRewards(suite.Ctx, period)
	})

	// Times are not stored for vaults with no state
	suite.StoredEarnTimeEquals(derivative0.Denom, time.Time{})
	suite.StoredEarnTimeEquals(derivative1.Denom, time.Time{})
	suite.StoredEarnIndexesEqual(derivative0.Denom, nil)
	suite.StoredEarnIndexesEqual(derivative1.Denom, nil)
}
