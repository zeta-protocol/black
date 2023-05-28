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
	EvmDenom = "afury"

	// CosmosDenom is the gas denom used by the black app
	CosmosDenom = "ublack"
)

// ConversionMultiplier is the conversion multiplier between afury and ublack
var ConversionMultiplier = sdkmath.NewInt(1_000_000_000_000)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal afury coin on the evm.
// x/evm consumes gas and send coins by minting and burning afury coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the ublack coin and a separate afury balance to manage the
// extra percision needed by the evm.
type EvmBankKeeper struct {
	afuryKeeper Keeper
	bk          types.BankKeeper
	ak          types.AccountKeeper
}

func NewEvmBankKeeper(afuryKeeper Keeper, bk types.BankKeeper, ak types.AccountKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		afuryKeeper: afuryKeeper,
		bk:          bk,
		ak:          ak,
	}
}

// GetBalance returns the total **spendable** balance of afury for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	ublack := spendableCoins.AmountOf(CosmosDenom)
	afury := k.afuryKeeper.GetBalance(ctx, addr)
	total := ublack.Mul(ConversionMultiplier).Add(afury)
	return sdk.NewCoin(EvmDenom, total)
}

// SendCoins transfers afury coins from a AccAddress to an AccAddress.
func (k EvmBankKeeper) SendCoins(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// SendCoins method is not used by the evm module, but is required by the
	// evmtypes.BankKeeper interface. This must be updated if the evm module
	// is updated to use SendCoins.
	panic("not implemented")
}

// SendCoinsFromModuleToAccount transfers afury coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (k EvmBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	ublack, afury, err := SplitAfuryCoins(amt)
	if err != nil {
		return err
	}

	if ublack.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(ublack)); err != nil {
			return err
		}
	}

	senderAddr := k.GetModuleAddress(senderModule)
	if err := k.ConvertOneUblackToAfuryIfNeeded(ctx, senderAddr, afury); err != nil {
		return err
	}

	if err := k.afuryKeeper.SendBalance(ctx, senderAddr, recipientAddr, afury); err != nil {
		return err
	}

	return k.ConvertAfuryToUblack(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers afury coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	ublack, afuryNeeded, err := SplitAfuryCoins(amt)
	if err != nil {
		return err
	}

	if ublack.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(ublack)); err != nil {
			return err
		}
	}

	if err := k.ConvertOneUblackToAfuryIfNeeded(ctx, senderAddr, afuryNeeded); err != nil {
		return err
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if err := k.afuryKeeper.SendBalance(ctx, senderAddr, recipientAddr, afuryNeeded); err != nil {
		return err
	}

	return k.ConvertAfuryToUblack(ctx, recipientAddr)
}

// MintCoins mints afury coins by minting the equivalent ublack coins and any remaining afury coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ublack, afury, err := SplitAfuryCoins(amt)
	if err != nil {
		return err
	}

	if ublack.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(ublack)); err != nil {
			return err
		}
	}

	recipientAddr := k.GetModuleAddress(moduleName)
	if err := k.afuryKeeper.AddBalance(ctx, recipientAddr, afury); err != nil {
		return err
	}

	return k.ConvertAfuryToUblack(ctx, recipientAddr)
}

// BurnCoins burns afury coins by burning the equivalent ublack coins and any remaining afury coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	ublack, afury, err := SplitAfuryCoins(amt)
	if err != nil {
		return err
	}

	if ublack.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(ublack)); err != nil {
			return err
		}
	}

	moduleAddr := k.GetModuleAddress(moduleName)
	if err := k.ConvertOneUblackToAfuryIfNeeded(ctx, moduleAddr, afury); err != nil {
		return err
	}

	return k.afuryKeeper.RemoveBalance(ctx, moduleAddr, afury)
}

// ConvertOneUblackToAfuryIfNeeded converts 1 ublack to afury for an address if
// its afury balance is smaller than the afuryNeeded amount.
func (k EvmBankKeeper) ConvertOneUblackToAfuryIfNeeded(ctx sdk.Context, addr sdk.AccAddress, afuryNeeded sdkmath.Int) error {
	afuryBal := k.afuryKeeper.GetBalance(ctx, addr)
	if afuryBal.GTE(afuryNeeded) {
		return nil
	}

	ublackToStore := sdk.NewCoins(sdk.NewCoin(CosmosDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, ublackToStore); err != nil {
		return err
	}

	// add 1ublack equivalent of afury to addr
	afuryToReceive := ConversionMultiplier
	if err := k.afuryKeeper.AddBalance(ctx, addr, afuryToReceive); err != nil {
		return err
	}

	return nil
}

// ConvertAfuryToUblack converts all available afury to ublack for a given AccAddress.
func (k EvmBankKeeper) ConvertAfuryToUblack(ctx sdk.Context, addr sdk.AccAddress) error {
	totalAfury := k.afuryKeeper.GetBalance(ctx, addr)
	ublack, _, err := SplitAfuryCoins(sdk.NewCoins(sdk.NewCoin(EvmDenom, totalAfury)))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough afury for a single ublack
	ublackToReceive := ublack.Amount
	if !ublackToReceive.IsPositive() {
		return nil
	}

	// remove afury used for converting to ublack
	afuryToBurn := ublackToReceive.Mul(ConversionMultiplier)
	finalBal := totalAfury.Sub(afuryToBurn)
	if err := k.afuryKeeper.SetBalance(ctx, addr, finalBal); err != nil {
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

// SplitAfuryCoins splits afury coins to the equivalent ublack coins and any remaining afury balance.
// An error will be returned if the coins are not valid or if the coins are not the afury denom.
func SplitAfuryCoins(coins sdk.Coins) (sdk.Coin, sdkmath.Int, error) {
	afury := sdk.ZeroInt()
	ublack := sdk.NewCoin(CosmosDenom, sdk.ZeroInt())

	if len(coins) == 0 {
		return ublack, afury, nil
	}

	if err := ValidateEvmCoins(coins); err != nil {
		return ublack, afury, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		afury = remainingBalance
	}
	ublackAmount := coin.Amount.Quo(ConversionMultiplier)
	if ublackAmount.IsPositive() {
		ublack = sdk.NewCoin(CosmosDenom, ublackAmount)
	}

	return ublack, afury, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the EvmDenom (afury).
func ValidateEvmCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is afury
	if len(coins) != 1 || coins[0].Denom != EvmDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", EvmDenom)
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
