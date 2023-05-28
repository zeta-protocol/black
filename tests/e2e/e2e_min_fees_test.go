package e2e_test

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/zeta-protocol/black/app"
	"github.com/zeta-protocol/black/tests/util"
)

func (suite *IntegrationTestSuite) TestEthGasPriceReturnsMinFee() {
	// read expected min fee from app.toml
	minGasPrices, err := getMinFeeFromAppToml(suite.BlackHomePath())
	suite.NoError(err)

	// evm uses afury, get afury min fee
	evmMinGas := minGasPrices.AmountOf("afury").TruncateInt().BigInt()

	// returns eth_gasPrice, units in black
	gasPrice, err := suite.Black.EvmClient.SuggestGasPrice(context.Background())
	suite.NoError(err)

	suite.Equal(evmMinGas, gasPrice)
}

func (suite *IntegrationTestSuite) TestEvmRespectsMinFee() {
	// setup sender & receiver
	sender := suite.Black.NewFundedAccount("evm-min-fee-test-sender", sdk.NewCoins(ufury(2e6)))
	randoReceiver := util.SdkToEvmAddress(app.RandomAddress())

	// get min gas price for evm (from app.toml)
	minFees, err := getMinFeeFromAppToml(suite.BlackHomePath())
	suite.NoError(err)
	minGasPrice := minFees.AmountOf("afury").TruncateInt()

	// attempt tx with less than min gas price (min fee - 1)
	tooLowGasPrice := minGasPrice.Sub(sdk.OneInt()).BigInt()
	req := util.EvmTxRequest{
		Tx:   ethtypes.NewTransaction(0, randoReceiver, big.NewInt(1e6), 1e5, tooLowGasPrice, nil),
		Data: "this tx should fail because it's gas price is too low",
	}
	res := sender.SignAndBroadcastEvmTx(req)

	// expect the tx to fail!
	suite.ErrorAs(res.Err, &util.ErrEvmFailedToBroadcast{})
	suite.ErrorContains(res.Err, "insufficient fee")
}

func getMinFeeFromAppToml(blackHome string) (sdk.DecCoins, error) {
	// read the expected min gas price from app.toml
	parsed := struct {
		MinGasPrices string `toml:"minimum-gas-prices"`
	}{}
	appToml, err := os.ReadFile(filepath.Join(blackHome, "config", "app.toml"))
	if err != nil {
		return nil, err
	}
	err = toml.Unmarshal(appToml, &parsed)
	if err != nil {
		return nil, err
	}

	// convert to dec coins
	return sdk.ParseDecCoins(strings.ReplaceAll(parsed.MinGasPrices, ";", ","))
}
