package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/zeta-protocol/black/x/evmutil/types"
)

const (
	// EvmDenom is the gas denom used by the evm
	EvmDenom = "ablack"

	// CosmosDenom is the gas denom used by the black app
	CosmosDenom = "ublack"
)

// ConversionMultiplier is the conversion multiplier between ablack and ublack
var ConversionMultiplier = sdkmath.NewInt(1_000_000_000_000)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal ablack coin on the evm.
// x/evm consumes gas and send coins by minting and burning ablack coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the ublack coin and a separate ablack balance to manage the
// extra percision needed by the evm.
type EvmBankKeeper struct {
	ablackKeeper Keeper
	bk          types.BankKeeper
	ak          types.AccountKeeper
}

func NewEvmBankKeeper(ablackKeeper Keeper, bk types.BankKeeper, ak types.AccountKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		ablackKeeper: ablackKeeper,
		bk:          bk,
		ak:          ak,
	}
}

// GetBalance returns the total **spendable** balance of ablack for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	ublack := spendableCoins.AmountOf(CosmosDenom)
	ablack := k.ablackKeeper.GetBalance(ctx, addr)
	total := ublack.Mul(ConversionMultiplier).Add(ablack)
	return sdk.NewCoin(EvmDenom, total)
}

// SendCoins transfers ablack coins from a AccAddress to an AccAddress.
func (k EvmBankKeeper) SendCoins(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// SendCoins method is not used by the evm module, but is required by the
	// evmtypes.BankKeeper interface. This must be updated if the evm module
	// is updated to use SendCoins.
	panic("not implemented")
}

// SendCoinsFromModuleToAccount transfers ablack coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (k EvmBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	ublack, ablack, err := SplitAblackCoins(amt)
	if err != nil {
		return err
	}

	if ublack.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(ublack)); err != nil {
			return err
		}
	}

	senderAddr := k.GetModuleAddress(senderModule)
	if err := k.ConvertOneUblackToAblackIfNeeded(ctx, senderAddr, ablack); err != nil {
		return err
	}

	if err := k.ablackKeeper.SendBalance(ctx, senderAddr, recipientAddr, ablack); err != nil {
		return err
	}

	return k.ConvertAblackToUblack(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers ablack coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	ublack, ablackNeeded, err := SplitAblackCoins(amt)
	if err != nil {
		return err
	}

	if ublack.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(ublack)); err != nil {
			return err
		}
	}

	if err := k.ConvertOneUblackToAblackIfNeeded(ctx, senderAddr, ablackNeeded); err != nil {
		return err
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if err := k.ablackKeeper.SendBalance(ctx, senderAddr, recipientAddr, ablackNeeded); err != nil {
		return err
	}

	return k.ConvertAblackToUblack(ctx, recipientAddr)
}

// MintCoins mints ablack coins by minting the equivalent ublack coins and any remaining ablack coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ublack, ablack, err := SplitAblackCoins(amt)
	if err != nil {
		return err
	}

	if ublack.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(ublack)); err != nil {
			return err
		}
	}

	recipientAddr := k.GetModuleAddress(moduleName)
	if err := k.ablackKeeper.AddBalance(ctx, recipientAddr, ablack); err != nil {
		return err
	}

	return k.ConvertAblackToUblack(ctx, recipientAddr)
}

// BurnCoins burns ablack coins by burning the equivalent ublack coins and any remaining ablack coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ublack, ablack, err := SplitAblackCoins(amt)
	if err != nil {
		return err
	}

	if ublack.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(ublack)); err != nil {
			return err
		}
	}

	moduleAddr := k.GetModuleAddress(moduleName)
	if err := k.ConvertOneUblackToAblackIfNeeded(ctx, moduleAddr, ablack); err != nil {
		return err
	}

	return k.ablackKeeper.RemoveBalance(ctx, moduleAddr, ablack)
}

// ConvertOneUblackToAblackIfNeeded converts 1 ublack to ablack for an address if
// its ablack balance is smaller than the ablackNeeded amount.
func (k EvmBankKeeper) ConvertOneUblackToAblackIfNeeded(ctx sdk.Context, addr sdk.AccAddress, ablackNeeded sdkmath.Int) error {
	ablackBal := k.ablackKeeper.GetBalance(ctx, addr)
	if ablackBal.GTE(ablackNeeded) {
		return nil
	}

	ublackToStore := sdk.NewCoins(sdk.NewCoin(CosmosDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, ublackToStore); err != nil {
		return err
	}

	// add 1ublack equivalent of ablack to addr
	ablackToReceive := ConversionMultiplier
	if err := k.ablackKeeper.AddBalance(ctx, addr, ablackToReceive); err != nil {
		return err
	}

	return nil
}

// ConvertAblackToUblack converts all available ablack to ublack for a given AccAddress.
func (k EvmBankKeeper) ConvertAblackToUblack(ctx sdk.Context, addr sdk.AccAddress) error {
	totalAblack := k.ablackKeeper.GetBalance(ctx, addr)
	ublack, _, err := SplitAblackCoins(sdk.NewCoins(sdk.NewCoin(EvmDenom, totalAblack)))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough ablack for a single ublack
	ublackToReceive := ublack.Amount
	if !ublackToReceive.IsPositive() {
		return nil
	}

	// remove ablack used for converting to ublack
	ablackToBurn := ublackToReceive.Mul(ConversionMultiplier)
	finalBal := totalAblack.Sub(ablackToBurn)
	if err := k.ablackKeeper.SetBalance(ctx, addr, finalBal); err != nil {
		return err
	}

	fromAddr := k.GetModuleAddress(types.ModuleName)
	if err := k.bk.SendCoins(ctx, fromAddr, addr, sdk.NewCoins(ublack)); err != nil {
		return err
	}

	return nil
}

func (k EvmBankKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	addr := k.ak.GetModuleAddress(moduleName)
	if addr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}
	return addr
}

// SplitAblackCoins splits ablack coins to the equivalent ublack coins and any remaining ablack balance.
// An error will be returned if the coins are not valid or if the coins are not the ablack denom.
func SplitAblackCoins(coins sdk.Coins) (sdk.Coin, sdkmath.Int, error) {
	ablack := sdk.ZeroInt()
	ublack := sdk.NewCoin(CosmosDenom, sdk.ZeroInt())

	if len(coins) == 0 {
		return ublack, ablack, nil
	}

	if err := ValidateEvmCoins(coins); err != nil {
		return ublack, ablack, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		ablack = remainingBalance
	}
	ublackAmount := coin.Amount.Quo(ConversionMultiplier)
	if ublackAmount.IsPositive() {
		ublack = sdk.NewCoin(CosmosDenom, ublackAmount)
	}

	return ublack, ablack, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the EvmDenom (ablack).
func ValidateEvmCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is ablack
	if len(coins) != 1 || coins[0].Denom != EvmDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", EvmDenom)
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
