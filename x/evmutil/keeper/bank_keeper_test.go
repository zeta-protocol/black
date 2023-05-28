package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmtime "github.com/tendermint/tendermint/types/time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/zeta-protocol/black/x/evmutil/keeper"
	"github.com/zeta-protocol/black/x/evmutil/testutil"
	"github.com/zeta-protocol/black/x/evmutil/types"
)

type evmBankKeeperTestSuite struct {
	testutil.Suite
}

func (suite *evmBankKeeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *evmBankKeeperTestSuite) TestGetBalance_ReturnsSpendable() {
	startingCoins := sdk.NewCoins(sdk.NewInt64Coin("ublack", 10))
	startingAfury := sdkmath.NewInt(100)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	bacc := authtypes.NewBaseAccountWithAddress(suite.Addrs[0])
	vacc := vesting.NewContinuousVestingAccount(bacc, startingCoins, now.Unix(), endTime.Unix())
	suite.AccountKeeper.SetAccount(suite.Ctx, vacc)

	err := suite.App.FundAccount(suite.Ctx, suite.Addrs[0], startingCoins)
	suite.Require().NoError(err)
	err = suite.Keeper.SetBalance(suite.Ctx, suite.Addrs[0], startingAfury)
	suite.Require().NoError(err)

	coin := suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "afury")
	suite.Require().Equal(startingAfury, coin.Amount)

	ctx := suite.Ctx.WithBlockTime(now.Add(12 * time.Hour))
	coin = suite.EvmBankKeeper.GetBalance(ctx, suite.Addrs[0], "afury")
	suite.Require().Equal(sdkmath.NewIntFromUint64(5_000_000_000_100), coin.Amount)
}

func (suite *evmBankKeeperTestSuite) TestGetBalance_NotEvmDenom() {
	suite.Require().Panics(func() {
		suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ublack")
	})
	suite.Require().Panics(func() {
		suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "busd")
	})
}

func (suite *evmBankKeeperTestSuite) TestGetBalance() {
	tests := []struct {
		name           string
		startingAmount sdk.Coins
		expAmount      sdkmath.Int
	}{
		{
			"ublack with afury",
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 100),
				sdk.NewInt64Coin("ublack", 10),
			),
			sdkmath.NewInt(10_000_000_000_100),
		},
		{
			"just afury",
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 100),
				sdk.NewInt64Coin("busd", 100),
			),
			sdkmath.NewInt(100),
		},
		{
			"just ublack",
			sdk.NewCoins(
				sdk.NewInt64Coin("ublack", 10),
				sdk.NewInt64Coin("busd", 100),
			),
			sdkmath.NewInt(10_000_000_000_000),
		},
		{
			"no ublack or afury",
			sdk.NewCoins(),
			sdk.ZeroInt(),
		},
		{
			"with avaka that is more than 1 ublack",
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 20_000_000_000_220),
				sdk.NewInt64Coin("ublack", 11),
			),
			sdkmath.NewInt(31_000_000_000_220),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithBlack(suite.Addrs[0], tt.startingAmount)
			coin := suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "afury")
			suite.Require().Equal(tt.expAmount, coin.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSendCoinsFromModuleToAccount() {
	startingModuleCoins := sdk.NewCoins(
		sdk.NewInt64Coin("afury", 200),
		sdk.NewInt64Coin("ublack", 100),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		startingAccBal sdk.Coins
		expAccBal      sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 12_000_000_000_010)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 10),
				sdk.NewInt64Coin("ublack", 12),
			),
			false,
		},
		{
			"send less than 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 122)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 122),
				sdk.NewInt64Coin("ublack", 0),
			),
			false,
		},
		{
			"send an exact amount of ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 98_000_000_000_000)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 0o0),
				sdk.NewInt64Coin("ublack", 98),
			),
			false,
		},
		{
			"send no afury",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 0)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 0),
				sdk.NewInt64Coin("ublack", 0),
			),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total afury to cover",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100_000_000_001_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough ublack to cover",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 200_000_000_000_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts receiver's afury to ublack if there's enough afury after the transfer",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 99_000_000_000_200)),
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 999_999_999_900),
				sdk.NewInt64Coin("ublack", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 100),
				sdk.NewInt64Coin("ublack", 101),
			),
			false,
		},
		{
			"converts all of receiver's afury to ublack even if somehow receiver has more than 1ublack of afury",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 12_000_000_000_100)),
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 5_999_999_999_990),
				sdk.NewInt64Coin("ublack", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 90),
				sdk.NewInt64Coin("ublack", 19),
			),
			false,
		},
		{
			"swap 1 ublack for afury if module account doesn't have enough afury",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 99_000_000_001_000)),
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 200),
				sdk.NewInt64Coin("ublack", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("afury", 1200),
				sdk.NewInt64Coin("ublack", 100),
			),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithBlack(suite.Addrs[0], tt.startingAccBal)
			suite.FundModuleAccountWithBlack(evmtypes.ModuleName, startingModuleCoins)

			// fund our module with some ublack to account for converting extra afury back to ublack
			suite.FundModuleAccountWithBlack(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("ublack", 10)))

			err := suite.EvmBankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, evmtypes.ModuleName, suite.Addrs[0], tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ublack
			ublackSender := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ublack")
			suite.Require().Equal(tt.expAccBal.AmountOf("ublack").Int64(), ublackSender.Amount.Int64())

			// check afury
			actualAfury := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expAccBal.AmountOf("afury").Int64(), actualAfury.Int64())
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSendCoinsFromAccountToModule() {
	startingAccCoins := sdk.NewCoins(
		sdk.NewInt64Coin("afury", 200),
		sdk.NewInt64Coin("ublack", 100),
	)
	startingModuleCoins := sdk.NewCoins(
		sdk.NewInt64Coin("afury", 100_000_000_000),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		expSenderCoins sdk.Coins
		expModuleCoins sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 12_000_000_000_010)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 190), sdk.NewInt64Coin("ublack", 88)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100_000_000_010), sdk.NewInt64Coin("ublack", 12)),
			false,
		},
		{
			"send less than 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 122)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 78), sdk.NewInt64Coin("ublack", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100_000_000_122), sdk.NewInt64Coin("ublack", 0)),
			false,
		},
		{
			"send an exact amount of ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 98_000_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 200), sdk.NewInt64Coin("ublack", 2)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100_000_000_000), sdk.NewInt64Coin("ublack", 98)),
			false,
		},
		{
			"send no afury",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 0)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 200), sdk.NewInt64Coin("ublack", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100_000_000_000), sdk.NewInt64Coin("ublack", 0)),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("afury", 12_000_000_000_000),
				sdk.NewInt64Coin("afury", 2_000_000_000_000),
			},
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total afury to cover",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100_000_000_001_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough ublack to cover",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 200_000_000_000_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts 1 ublack to afury if not enough afury to cover",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 99_001_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 999_000_000_200), sdk.NewInt64Coin("ublack", 0)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 101_000_000_000), sdk.NewInt64Coin("ublack", 99)),
			false,
		},
		{
			"converts receiver's afury to ublack if there's enough afury after the transfer",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 5_900_000_000_200)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100_000_000_000), sdk.NewInt64Coin("ublack", 94)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 200), sdk.NewInt64Coin("ublack", 6)),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			suite.FundAccountWithBlack(suite.Addrs[0], startingAccCoins)
			suite.FundModuleAccountWithBlack(evmtypes.ModuleName, startingModuleCoins)

			err := suite.EvmBankKeeper.SendCoinsFromAccountToModule(suite.Ctx, suite.Addrs[0], evmtypes.ModuleName, tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check sender balance
			ublackSender := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ublack")
			suite.Require().Equal(tt.expSenderCoins.AmountOf("ublack").Int64(), ublackSender.Amount.Int64())
			actualAfury := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expSenderCoins.AmountOf("afury").Int64(), actualAfury.Int64())

			// check module balance
			moduleAddr := suite.AccountKeeper.GetModuleAddress(evmtypes.ModuleName)
			ublackSender = suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, "ublack")
			suite.Require().Equal(tt.expModuleCoins.AmountOf("ublack").Int64(), ublackSender.Amount.Int64())
			actualAfury = suite.Keeper.GetBalance(suite.Ctx, moduleAddr)
			suite.Require().Equal(tt.expModuleCoins.AmountOf("afury").Int64(), actualAfury.Int64())
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestBurnCoins() {
	startingUblack := sdkmath.NewInt(100)
	tests := []struct {
		name       string
		burnCoins  sdk.Coins
		expUblack   sdkmath.Int
		expAfury   sdkmath.Int
		hasErr     bool
		afuryStart sdkmath.Int
	}{
		{
			"burn more than 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 12_021_000_000_002)),
			sdkmath.NewInt(88),
			sdkmath.NewInt(100_000_000_000),
			false,
			sdkmath.NewInt(121_000_000_002),
		},
		{
			"burn less than 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 122)),
			sdkmath.NewInt(100),
			sdkmath.NewInt(878),
			false,
			sdkmath.NewInt(1000),
		},
		{
			"burn an exact amount of ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 98_000_000_000_000)),
			sdkmath.NewInt(2),
			sdkmath.NewInt(10),
			false,
			sdkmath.NewInt(10),
		},
		{
			"burn no afury",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 0)),
			startingUblack,
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if burning other coins",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 500), sdk.NewInt64Coin("busd", 1000)),
			startingUblack,
			sdkmath.NewInt(100),
			true,
			sdkmath.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("afury", 12_000_000_000_000),
				sdk.NewInt64Coin("afury", 2_000_000_000_000),
			},
			startingUblack,
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if burn amount is negative",
			sdk.Coins{sdk.Coin{Denom: "afury", Amount: sdkmath.NewInt(-100)}},
			startingUblack,
			sdkmath.NewInt(50),
			true,
			sdkmath.NewInt(50),
		},
		{
			"errors if not enough afury to cover burn",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100_999_000_000_000)),
			sdkmath.NewInt(0),
			sdkmath.NewInt(99_000_000_000),
			true,
			sdkmath.NewInt(99_000_000_000),
		},
		{
			"errors if not enough ublack to cover burn",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 200_000_000_000_000)),
			sdkmath.NewInt(100),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"converts 1 ublack to afury if not enough afury to cover",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 12_021_000_000_002)),
			sdkmath.NewInt(87),
			sdkmath.NewInt(980_000_000_000),
			false,
			sdkmath.NewInt(1_000_000_002),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			startingCoins := sdk.NewCoins(
				sdk.NewCoin("ublack", startingUblack),
				sdk.NewCoin("afury", tt.afuryStart),
			)
			suite.FundModuleAccountWithBlack(evmtypes.ModuleName, startingCoins)

			err := suite.EvmBankKeeper.BurnCoins(suite.Ctx, evmtypes.ModuleName, tt.burnCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ublack
			ublackActual := suite.BankKeeper.GetBalance(suite.Ctx, suite.EvmModuleAddr, "ublack")
			suite.Require().Equal(tt.expUblack, ublackActual.Amount)

			// check afury
			afuryActual := suite.Keeper.GetBalance(suite.Ctx, suite.EvmModuleAddr)
			suite.Require().Equal(tt.expAfury, afuryActual)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestMintCoins() {
	tests := []struct {
		name       string
		mintCoins  sdk.Coins
		ublack      sdkmath.Int
		afury      sdkmath.Int
		hasErr     bool
		afuryStart sdkmath.Int
	}{
		{
			"mint more than 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 12_021_000_000_002)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(21_000_000_002),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint less than 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 901_000_000_001)),
			sdk.ZeroInt(),
			sdkmath.NewInt(901_000_000_001),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint an exact amount of ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 123_000_000_000_000_000)),
			sdkmath.NewInt(123_000),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint no afury",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 0)),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if minting other coins",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.ZeroInt(),
			sdkmath.NewInt(100),
			true,
			sdkmath.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("afury", 12_000_000_000_000),
				sdk.NewInt64Coin("afury", 2_000_000_000_000),
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if mint amount is negative",
			sdk.Coins{sdk.Coin{Denom: "afury", Amount: sdkmath.NewInt(-100)}},
			sdk.ZeroInt(),
			sdkmath.NewInt(50),
			true,
			sdkmath.NewInt(50),
		},
		{
			"adds to existing afury balance",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 12_021_000_000_002)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(21_000_000_102),
			false,
			sdkmath.NewInt(100),
		},
		{
			"convert afury balance to ublack if it exceeds 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 10_999_000_000_000)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(1_200_000_001),
			false,
			sdkmath.NewInt(1_002_200_000_001),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			suite.FundModuleAccountWithBlack(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("ublack", 10)))
			suite.FundModuleAccountWithBlack(evmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin("afury", tt.afuryStart)))

			err := suite.EvmBankKeeper.MintCoins(suite.Ctx, evmtypes.ModuleName, tt.mintCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check ublack
			ublackActual := suite.BankKeeper.GetBalance(suite.Ctx, suite.EvmModuleAddr, "ublack")
			suite.Require().Equal(tt.ublack, ublackActual.Amount)

			// check afury
			afuryActual := suite.Keeper.GetBalance(suite.Ctx, suite.EvmModuleAddr)
			suite.Require().Equal(tt.afury, afuryActual)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestValidateEvmCoins() {
	tests := []struct {
		name      string
		coins     sdk.Coins
		shouldErr bool
	}{
		{
			"valid coins",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 500)),
			false,
		},
		{
			"dup coins",
			sdk.Coins{sdk.NewInt64Coin("afury", 500), sdk.NewInt64Coin("afury", 500)},
			true,
		},
		{
			"not evm coins",
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 500)),
			true,
		},
		{
			"negative coins",
			sdk.Coins{sdk.Coin{Denom: "afury", Amount: sdkmath.NewInt(-500)}},
			true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := keeper.ValidateEvmCoins(tt.coins)
			if tt.shouldErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestConvertOneUblackToAfuryIfNeeded() {
	afuryNeeded := sdkmath.NewInt(200)
	tests := []struct {
		name          string
		startingCoins sdk.Coins
		expectedCoins sdk.Coins
		success       bool
	}{
		{
			"not enough ublack for conversion",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100)),
			false,
		},
		{
			"converts 1 ublack to afury",
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 10), sdk.NewInt64Coin("afury", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 9), sdk.NewInt64Coin("afury", 1_000_000_000_100)),
			true,
		},
		{
			"conversion not needed",
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 10), sdk.NewInt64Coin("afury", 200)),
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 10), sdk.NewInt64Coin("afury", 200)),
			true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithBlack(suite.Addrs[0], tt.startingCoins)
			err := suite.EvmBankKeeper.ConvertOneUblackToAfuryIfNeeded(suite.Ctx, suite.Addrs[0], afuryNeeded)
			moduleBlack := suite.BankKeeper.GetBalance(suite.Ctx, suite.AccountKeeper.GetModuleAddress(types.ModuleName), "ublack")
			if tt.success {
				suite.Require().NoError(err)
				if tt.startingCoins.AmountOf("afury").LT(afuryNeeded) {
					suite.Require().Equal(sdk.OneInt(), moduleBlack.Amount)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().Equal(sdk.ZeroInt(), moduleBlack.Amount)
			}

			afury := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expectedCoins.AmountOf("afury"), afury)
			ublack := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ublack")
			suite.Require().Equal(tt.expectedCoins.AmountOf("ublack"), ublack.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestConvertAfuryToUblack() {
	tests := []struct {
		name          string
		startingCoins sdk.Coins
		expectedCoins sdk.Coins
	}{
		{
			"not enough ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 100), sdk.NewInt64Coin("ublack", 0)),
		},
		{
			"converts afury for 1 ublack",
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 10), sdk.NewInt64Coin("afury", 1_000_000_000_003)),
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 11), sdk.NewInt64Coin("afury", 3)),
		},
		{
			"converts more than 1 ublack of afury",
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 10), sdk.NewInt64Coin("afury", 8_000_000_000_123)),
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 18), sdk.NewInt64Coin("afury", 123)),
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("ublack", 10)))
			suite.Require().NoError(err)
			suite.FundAccountWithBlack(suite.Addrs[0], tt.startingCoins)
			err = suite.EvmBankKeeper.ConvertAfuryToUblack(suite.Ctx, suite.Addrs[0])
			suite.Require().NoError(err)
			afury := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expectedCoins.AmountOf("afury"), afury)
			ublack := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "ublack")
			suite.Require().Equal(tt.expectedCoins.AmountOf("ublack"), ublack.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSplitAfuryCoins() {
	tests := []struct {
		name          string
		coins         sdk.Coins
		expectedCoins sdk.Coins
		shouldErr     bool
	}{
		{
			"invalid coins",
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 500)),
			nil,
			true,
		},
		{
			"empty coins",
			sdk.NewCoins(),
			sdk.NewCoins(),
			false,
		},
		{
			"ublack & afury coins",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 8_000_000_000_123)),
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 8), sdk.NewInt64Coin("afury", 123)),
			false,
		},
		{
			"only afury",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 10_123)),
			sdk.NewCoins(sdk.NewInt64Coin("afury", 10_123)),
			false,
		},
		{
			"only ublack",
			sdk.NewCoins(sdk.NewInt64Coin("afury", 5_000_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin("ublack", 5)),
			false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ublack, afury, err := keeper.SplitAfuryCoins(tt.coins)
			if tt.shouldErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tt.expectedCoins.AmountOf("ublack"), ublack.Amount)
				suite.Require().Equal(tt.expectedCoins.AmountOf("afury"), afury)
			}
		})
	}
}

func TestEvmBankKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(evmBankKeeperTestSuite))
}
