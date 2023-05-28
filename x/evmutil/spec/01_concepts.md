<!--
order: 1
-->

# Concepts

## EVM Gas Denom

In order to use the EVM and be compatible with existing clients, the gas denom used by the EVM must be in 18 decimals. Since `ufury` has 6 decimals of precision, it cannot be used as the EVM gas denom directly.

To use the Black token on the EVM, the evmutil module provides an `EvmBankKeeper` that is responsible for the conversion of `ufury` and `afury`. A user's excess `afury` balance is stored in the `x/evmutil` store, while its `ufury` balance remains in the cosmos-sdk `x/bank` module.

## `EvmBankKeeper` Overview

The `EvmBankKeeper` provides access to an account's total `afury` balance and the ability to transfer, mint, and burn `afury`. If anything other than the `afury` denom is requested, the `EvmBankKeeper` will panic.

This keeper implements the `x/evm` module's `BankKeeper` interface to enable the usage of `afury` denom on the EVM.

### `x/evm` Parameter Requirements

Since the EVM denom `afury` is required to use the `EvmBankKeeper`, it is necessary to set the `EVMDenom` param of the `x/evm` module to `afury`.

### Balance Calculation of `afury`

The `afury` balance of an account is derived from an account's **spendable** `ufury` balance times 10^12 (to derive its `afury` equivalent), plus the account's excess `afury` balance that can be accessed via the module `Keeper`.

### `afury` <> `ufury` Conversion

When an account does not have sufficient `afury` to cover a transfer or burn, the `EvmBankKeeper` will try to swap 1 `ufury` to its equivalent `afury` amount. It does this by transferring 1 `ufury` from the sender to the `x/evmutil` module account, then adding the equivalent `afury` amount to the sender's balance in the module state.

In reverse, if an account has enough `afury` balance for one or more `ufury`, the excess `afury` balance will be converted to `ufury`. This is done by removing the excess `afury` balance in the module store, then transferring the equivalent `ufury` coins from the `x/evmutil` module account to the target account.

The swap logic ensures that all `afury` is backed by the equivalent `ufury` balance stored in the module account.

## ERC20 token <> sdk.Coin Conversion

`x/evmutil` enables the conversion between ERC20 tokens and sdk.Coins. This done through the use of the `MsgConvertERC20ToCoin` & `MsgConvertCoinToERC20` messages (see **[Messages](03_messages.md)**).

Only ERC20 contract address that are whitelist via the `EnabledConversionPairs` param (see **[Params](05_params.md)**) can be converted via these messages.

## Module Keeper

The module Keeper provides access to an account's excess `afury` balance and the ability to update the balance.
