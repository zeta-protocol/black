package e2e_test

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	emtypes "github.com/evmos/ethermint/types"

	"github.com/zeta-protocol/black/app"
	"github.com/zeta-protocol/black/tests/e2e/testutil"
	"github.com/zeta-protocol/black/tests/util"
)

var (
	minEvmGasPrice = big.NewInt(1e10) // afury
)

func ublack(amt int64) sdk.Coin {
	return sdk.NewCoin("ublack", sdkmath.NewInt(amt))
}

type IntegrationTestSuite struct {
	testutil.E2eTestSuite
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// example test that queries black via SDK and EVM
func (suite *IntegrationTestSuite) TestChainID() {
	expectedEvmNetworkId, err := emtypes.ParseChainID(suite.Black.ChainId)
	suite.NoError(err)

	// EVM query
	evmNetworkId, err := suite.Black.EvmClient.NetworkID(context.Background())
	suite.NoError(err)
	suite.Equal(expectedEvmNetworkId, evmNetworkId)

	// SDK query
	nodeInfo, err := suite.Black.Tm.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
	suite.NoError(err)
	suite.Equal(suite.Black.ChainId, nodeInfo.DefaultNodeInfo.Network)
}

// example test that funds a new account & queries its balance
func (suite *IntegrationTestSuite) TestFundedAccount() {
	funds := ublack(1e7)
	acc := suite.Black.NewFundedAccount("example-acc", sdk.NewCoins(funds))

	// check that the sdk & evm signers are for the same account
	suite.Equal(acc.SdkAddress.String(), util.EvmToSdkAddress(acc.EvmAddress).String())
	suite.Equal(acc.EvmAddress.Hex(), util.SdkToEvmAddress(acc.SdkAddress).Hex())

	// check balance via SDK query
	res, err := suite.Black.Bank.Balance(context.Background(), banktypes.NewQueryBalanceRequest(
		acc.SdkAddress, "ublack",
	))
	suite.NoError(err)
	suite.Equal(funds, *res.Balance)

	// check balance via EVM query
	afuryBal, err := suite.Black.EvmClient.BalanceAt(context.Background(), acc.EvmAddress, nil)
	suite.NoError(err)
	suite.Equal(funds.Amount.MulRaw(1e12).BigInt(), afuryBal)
}

// example test that signs & broadcasts an EVM tx
func (suite *IntegrationTestSuite) TestTransferOverEVM() {
	// fund an account that can perform the transfer
	initialFunds := ublack(1e7) // 10 BLACK
	acc := suite.Black.NewFundedAccount("evm-test-transfer", sdk.NewCoins(initialFunds))

	// get a rando account to send black to
	randomAddr := app.RandomAddress()
	to := util.SdkToEvmAddress(randomAddr)

	// example fetching of nonce (account sequence)
	nonce, err := suite.Black.EvmClient.PendingNonceAt(context.Background(), acc.EvmAddress)
	suite.NoError(err)
	suite.Equal(uint64(0), nonce) // sanity check. the account should have no prior txs

	// transfer black over EVM
	blackToTransfer := big.NewInt(1e18) // 1 BLACK; afury has 18 decimals.
	req := util.EvmTxRequest{
		Tx:   ethtypes.NewTransaction(nonce, to, blackToTransfer, 1e5, minEvmGasPrice, nil),
		Data: "any ol' data to track this through the system",
	}
	res := acc.SignAndBroadcastEvmTx(req)
	suite.NoError(res.Err)
	suite.Equal(ethtypes.ReceiptStatusSuccessful, res.Receipt.Status)

	// evm txs refund unused gas. so to know the expected balance we need to know how much gas was used.
	ublackUsedForGas := sdkmath.NewIntFromBigInt(minEvmGasPrice).
		Mul(sdkmath.NewIntFromUint64(res.Receipt.GasUsed)).
		QuoRaw(1e12) // convert afury to ublack

	// expect (9 - gas used) BLACK remaining in account.
	balance := suite.Black.QuerySdkForBalances(acc.SdkAddress)
	suite.Equal(sdkmath.NewInt(9e6).Sub(ublackUsedForGas), balance.AmountOf("ublack"))
}

// TestIbcTransfer transfers BLACK from the primary black chain (suite.Black) to the ibc chain (suite.Ibc).
// Note that because the IBC chain also runs black's binary, this tests both the sending & receiving.
func (suite *IntegrationTestSuite) TestIbcTransfer() {
	suite.SkipIfIbcDisabled()

	// ARRANGE
	// setup black account
	funds := ublack(1e7) // 10 BLACK
	blackAcc := suite.Black.NewFundedAccount("ibc-transfer-black-side", sdk.NewCoins(funds))
	// setup ibc account
	ibcAcc := suite.Ibc.NewFundedAccount("ibc-transfer-ibc-side", sdk.NewCoins())

	gasLimit := int64(2e5)
	fee := ublack(7500)

	fundsToSend := ublack(5e6) // 5 BLACK
	transferMsg := ibctypes.NewMsgTransfer(
		testutil.IbcPort,
		testutil.IbcChannel,
		fundsToSend,
		blackAcc.SdkAddress.String(),
		ibcAcc.SdkAddress.String(),
		ibcclienttypes.NewHeight(0, 0), // timeout height disabled when 0
		uint64(time.Now().Add(30*time.Second).UnixNano()),
		"",
	)
	// initial - sent - fee
	expectedSrcBalance := funds.Sub(fundsToSend).Sub(fee)

	// ACT
	// IBC transfer from black -> ibc
	transferTo := util.BlackMsgRequest{
		Msgs:      []sdk.Msg{transferMsg},
		GasLimit:  uint64(gasLimit),
		FeeAmount: sdk.NewCoins(fee),
		Memo:      "sent from Black!",
	}
	res := blackAcc.SignAndBroadcastBlackTx(transferTo)

	// ASSERT
	suite.NoError(res.Err)

	// the balance should be deducted from black account
	suite.Eventually(func() bool {
		balance := suite.Black.QuerySdkForBalances(blackAcc.SdkAddress)
		return balance.AmountOf("ublack").Equal(expectedSrcBalance.Amount)
	}, 10*time.Second, 1*time.Second)

	// expect the balance to be transferred to the ibc chain!
	suite.Eventually(func() bool {
		balance := suite.Ibc.QuerySdkForBalances(ibcAcc.SdkAddress)
		found := false
		for _, c := range balance {
			// find the ibc denom coin
			if strings.HasPrefix(c.Denom, "ibc/") {
				suite.Equal(fundsToSend.Amount, c.Amount)
				found = true
			}
		}
		return found
	}, 15*time.Second, 1*time.Second)
}
