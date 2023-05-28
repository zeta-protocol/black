package testutil

import (
	"fmt"
	"math/big"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	"github.com/zeta-protocol/black/app"
	"github.com/zeta-protocol/black/tests/e2e/runner"
)

const (
	FundedAccountName = "whale"
	// use coin type 60 so we are compatible with accounts from `black add keys --eth <name>`
	// these accounts use the ethsecp256k1 signing algorithm that allows the signing client
	// to manage both sdk & evm txs.
	Bip44CoinType = 60

	IbcPort    = "transfer"
	IbcChannel = "channel-0"
)

type E2eTestSuite struct {
	suite.Suite

	config SuiteConfig
	runner runner.NodeRunner

	Black *Chain
	Ibc  *Chain

	UpgradeHeight        int64
	DeployedErc20Address common.Address
}

func (suite *E2eTestSuite) SetupSuite() {
	var err error
	fmt.Println("setting up test suite.")
	app.SetSDKConfig()

	suiteConfig := ParseSuiteConfig()
	suite.config = suiteConfig
	suite.UpgradeHeight = suiteConfig.BlackUpgradeHeight
	suite.DeployedErc20Address = common.HexToAddress(suiteConfig.BlackErc20Address)

	runnerConfig := runner.Config{
		BlackConfigTemplate: suiteConfig.BlackConfigTemplate,

		IncludeIBC: suiteConfig.IncludeIbcTests,
		ImageTag:   "local",

		EnableAutomatedUpgrade:  suiteConfig.IncludeAutomatedUpgrade,
		BlackUpgradeName:         suiteConfig.BlackUpgradeName,
		BlackUpgradeHeight:       suiteConfig.BlackUpgradeHeight,
		BlackUpgradeBaseImageTag: suiteConfig.BlackUpgradeBaseImageTag,

		SkipShutdown: suiteConfig.SkipShutdown,
	}
	suite.runner = runner.NewBlackNode(runnerConfig)

	chains := suite.runner.StartChains()
	blackchain := chains.MustGetChain("black")
	suite.Black, err = NewChain(suite.T(), blackchain, suiteConfig.FundedAccountMnemonic)
	if err != nil {
		suite.runner.Shutdown()
		suite.T().Fatalf("failed to create black chain querier: %s", err)
	}

	if suiteConfig.IncludeIbcTests {
		ibcchain := chains.MustGetChain("ibc")
		suite.Ibc, err = NewChain(suite.T(), ibcchain, suiteConfig.FundedAccountMnemonic)
		if err != nil {
			suite.runner.Shutdown()
			suite.T().Fatalf("failed to create ibc chain querier: %s", err)
		}
	}

	suite.InitBlackEvmData()
}

func (suite *E2eTestSuite) TearDownSuite() {
	fmt.Println("tearing down test suite.")
	// close all account request channels
	suite.Black.Shutdown()
	if suite.Ibc != nil {
		suite.Ibc.Shutdown()
	}
	// gracefully shutdown docker container(s)
	suite.runner.Shutdown()
}

func (suite *E2eTestSuite) SkipIfIbcDisabled() {
	if !suite.config.IncludeIbcTests {
		suite.T().SkipNow()
	}
}

func (suite *E2eTestSuite) SkipIfUpgradeDisabled() {
	if !suite.config.IncludeAutomatedUpgrade {
		suite.T().SkipNow()
	}
}

// BlackHomePath returns the OS-specific filepath for the black home directory
// Assumes network is running with kvtool installed from the sub-repository in tests/e2e/kvtool
func (suite *E2eTestSuite) BlackHomePath() string {
	return filepath.Join("kvtool", "full_configs", "generated", "black", "initstate", ".black")
}

// BigIntsEqual is a helper method for comparing the equality of two big ints
func (suite *E2eTestSuite) BigIntsEqual(expected *big.Int, actual *big.Int, msg string) {
	suite.Truef(expected.Cmp(actual) == 0, "%s (expected: %s, actual: %s)", msg, expected.String(), actual.String())
}
