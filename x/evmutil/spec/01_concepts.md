<!--
order: 1
-->

# Concepts

## EVM Gas Denom

In order to use the EVM and be compatible with existing clients, the gas denom used by the EVM must be in 18 decimals. Since `ublack` has 6 decimals of precision, it cannot be used as the EVM gas denom directly.

To use the Black token on the EVM, the evmutil module provides an `EvmBankKeeper` that is responsible for the conversion of `ublack` and `ablack`. A user's excess `ablack` balance is stored in the `x/evmutil` store, while its `ublack` balance remains in the cosmos-sdk `x/bank` module.

## `EvmBankKeeper` Overview

The `EvmBankKeeper` provides access to an account's total `ablack` balance and the ability to transfer, mint, and burn `ablack`. If anything other than the `ablack` denom is requested, the `EvmBankKeeper` will panic.

This keeper implements the `x/evm` module's `BankKeeper` interface to enable the usage of `ablack` denom on the EVM.

### `x/evm` Parameter Requirements

Since the EVM denom `ablack` is required to use the `EvmBankKeeper`, it is necessary to set the `EVMDenom` param of the `x/evm` module to `ablack`.

### Balance Calculation of `ablack`

The `ablack` balance of an account is derived from an account's **spendable** `ublack` balance times 10^12 (to derive its `ablack` equivalent), plus the account's excess `ablack` balance that can be accessed via the module `Keeper`.

### `ablack` <> `ublack` Conversion

When an account does not have sufficient `ablack` to cover a transfer or burn, the `EvmBankKeeper` will try to swap 1 `ublack` to its equivalent `ablack` amount. It does this by transferring 1 `ublack` from the sender to the `x/evmutil` module account, then adding the equivalent `ablack` amount to the sender's balance in the module state.

In reverse, if an account has enough `ablack` balance for one or more `ublack`, the excess `ablack` balance will be converted to `ublack`. This is done by removing the excess `ablack` balance in the module store, then transferring the equivalent `ublack` coins from the `x/evmutil` module account to the target account.

The swap logic ensures that all `ablack` is backed by the equivalent `ublack` balance stored in the module account.

## ERC20 token <> sdk.Coin Conversion

`x/evmutil` enables the conversion between ERC20 tokens and sdk.Coins. This done through the use of the `MsgConvertERC20ToCoin` & `MsgConvertCoinToERC20` messages (see **[Messages](03_messages.md)**).

Only ERC20 contract address that are whitelist via the `EnabledConversionPairs` param (see **[Params](05_params.md)**) can be converted via these messages.

## Module Keeper

The module Keeper provides access to an account's excess `ablack` balance and the ability to update the balance.
