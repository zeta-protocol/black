package clitest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmtypes "github.com/tendermint/tendermint/types"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/zeta-protocol/black/app"
)

const (
	denom        = "stake"
	keyFoo       = "foo"
	keyBar       = "bar"
	fooDenom     = "footoken"
	feeDenom     = "feetoken"
	fee2Denom    = "fee2token"
	keyBaz       = "baz"
	keyVesting   = "vesting"
	keyFooBarBaz = "foobarbaz"
)

var (
	totalCoins = sdk.NewCoins(
		sdk.NewCoin(fee2Denom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(2000000)),
		sdk.NewCoin(fooDenom, sdk.TokensFromConsensusPower(2000)),
		sdk.NewCoin(denom, sdk.TokensFromConsensusPower(300).Add(sdkmath.NewInt(12))), // add coins from inflation
	)

	startCoins = sdk.NewCoins(
		sdk.NewCoin(fee2Denom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(1000000)),
		sdk.NewCoin(fooDenom, sdk.TokensFromConsensusPower(1000)),
		sdk.NewCoin(denom, sdk.TokensFromConsensusPower(150)),
	)

	vestingCoins = sdk.NewCoins(
		sdk.NewCoin(feeDenom, sdk.TokensFromConsensusPower(500000)),
	)
)

func init() {
	// set the address prefixes
	config := sdk.GetConfig()
	app.SetBech32AddressPrefixes(config)
	// config.Seal()
}

//___________________________________________________________________________________
// Fixtures

// Fixtures is used to setup the testing environment
type Fixtures struct {
	BuildDir    string
	RootDir     string
	BkdBinary   string
	BkcliBinary string
	ChainID     string
	RPCAddr     string
	Port        string
	BkdHome     string
	BkcliHome   string
	P2PAddr     string
	T           *testing.T

	cdc *codec.Codec
}

// NewFixtures creates a new instance of Fixtures with many vars set
func NewFixtures(t *testing.T) *Fixtures {
	tmpDir, err := ioutil.TempDir("", "black_integration_"+t.Name()+"_")
	require.NoError(t, err)

	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)

	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)

	buildDir := os.Getenv("BUILDDIR")
	if buildDir == "" {
		buildDir, err = filepath.Abs(filepath.Join("../build/", runtime.GOOS))
		require.NoError(t, err)
	}

	cdc := app.MakeCodec()

	return &Fixtures{
		T:           t,
		BuildDir:    buildDir,
		RootDir:     tmpDir,
		BkdBinary:   filepath.Join(buildDir, "bkd"),
		BkcliBinary: filepath.Join(buildDir, "bkcli"),
		BkdHome:     filepath.Join(tmpDir, ".bkd"),
		BkcliHome:   filepath.Join(tmpDir, ".bkcli"),
		RPCAddr:     servAddr,
		P2PAddr:     p2pAddr,
		Port:        port,
		cdc:         cdc,
	}
}

// GenesisFile returns the path of the genesis file
func (f Fixtures) GenesisFile() string {
	return filepath.Join(f.BkdHome, "config", "genesis.json")
}

// GenesisState returns the application's genesis state
func (f Fixtures) GenesisState() simapp.GenesisState {
	cdc := codec.New()
	genDoc, err := tmtypes.GenesisDocFromFile(f.GenesisFile())
	require.NoError(f.T, err)

	var appState simapp.GenesisState
	require.NoError(f.T, cdc.UnmarshalJSON(genDoc.AppState, &appState))
	return appState
}

// InitFixtures is called at the beginning of a test  and initializes a chain
// with 1 validator.
func InitFixtures(t *testing.T) (f *Fixtures) {
	f = NewFixtures(t)

	// reset test state
	f.UnsafeResetAll()

	f.CLIConfig("keyring-backend", "test")

	// ensure keystore has foo and bar keys
	f.KeysDelete(keyFoo)
	f.KeysDelete(keyBar)
	f.KeysDelete(keyBar)
	f.KeysDelete(keyFooBarBaz)
	f.KeysAdd(keyFoo)
	f.KeysAdd(keyBar)
	f.KeysAdd(keyBaz)
	f.KeysAdd(keyVesting)
	f.KeysAdd(keyFooBarBaz, "--multisig-threshold=2", fmt.Sprintf(
		"--multisig=%s,%s,%s", keyFoo, keyBar, keyBaz))

	// ensure that CLI output is in JSON format
	f.CLIConfig("output", "json")

	// NOTE: KvInit sets the ChainID
	f.KvInit(keyFoo)

	f.CLIConfig("chain-id", f.ChainID)
	f.CLIConfig("broadcast-mode", "block")
	f.CLIConfig("trust-node", "true")

	// start an account with tokens
	f.AddGenesisAccount(f.KeyAddress(keyFoo), startCoins)
	f.AddGenesisAccount(
		f.KeyAddress(keyVesting), startCoins,
		fmt.Sprintf("--vesting-amount=%s", vestingCoins),
		fmt.Sprintf("--vesting-start-time=%d", time.Now().UTC().UnixNano()),
		fmt.Sprintf("--vesting-end-time=%d", time.Now().Add(60*time.Second).UTC().UnixNano()),
	)

	f.GenTx(keyFoo)
	f.CollectGenTxs()

	return
}

// Cleanup is meant to be run at the end of a test to clean up an remaining test state
func (f *Fixtures) Cleanup(dirs ...string) {
	clean := append(dirs, f.RootDir)
	for _, d := range clean {
		require.NoError(f.T, os.RemoveAll(d))
	}
}

// Flags returns the flags necessary for making most CLI calls
func (f *Fixtures) Flags() string {
	return fmt.Sprintf("--home=%s --node=%s", f.BkcliHome, f.RPCAddr)
}

//___________________________________________________________________________________
// blackd

// UnsafeResetAll is blackd unsafe-reset-all
func (f *Fixtures) UnsafeResetAll(flags ...string) {
	cmd := fmt.Sprintf("%s --home=%s unsafe-reset-all", f.BkdBinary, f.BkdHome)
	executeWrite(f.T, addFlags(cmd, flags))
	err := os.RemoveAll(filepath.Join(f.BkdHome, "config", "gentx"))
	require.NoError(f.T, err)
}

// KvInit is blackd init
// NOTE: KvInit sets the ChainID for the Fixtures instance
func (f *Fixtures) KvInit(moniker string, flags ...string) {
	cmd := fmt.Sprintf("%s init -o --home=%s %s", f.BkdBinary, f.BkdHome, moniker)
	_, stderr := tests.ExecuteT(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)

	var chainID string
	var initRes map[string]json.RawMessage

	err := json.Unmarshal([]byte(stderr), &initRes)
	require.NoError(f.T, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(f.T, err)

	f.ChainID = chainID
}

// AddGenesisAccount is blackd add-genesis-account
func (f *Fixtures) AddGenesisAccount(address sdk.AccAddress, coins sdk.Coins, flags ...string) {
	cmd := fmt.Sprintf("%s add-genesis-account %s %s --home=%s --keyring-backend=test", f.BkdBinary, address, coins, f.BkdHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

// GenTx is blackd gentx
func (f *Fixtures) GenTx(name string, flags ...string) {
	cmd := fmt.Sprintf("%s gentx --name=%s --home=%s --home-client=%s --keyring-backend=test", f.BkdBinary, name, f.BkdHome, f.BkcliHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

// CollectGenTxs is blackd collect-gentxs
func (f *Fixtures) CollectGenTxs(flags ...string) {
	cmd := fmt.Sprintf("%s collect-gentxs --home=%s", f.BkdBinary, f.BkdHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

// GDStart runs blackd start with the appropriate flags and returns a process
func (f *Fixtures) GDStart(flags ...string) *tests.Process {
	cmd := fmt.Sprintf("%s start --home=%s --rpc.laddr=%v --p2p.laddr=%v --pruning=everything", f.BkdBinary, f.BkdHome, f.RPCAddr, f.P2PAddr)
	proc := tests.GoExecuteTWithStdout(f.T, addFlags(cmd, flags))
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)
	return proc
}

// GDTendermint returns the results of blackd tendermint [query]
func (f *Fixtures) GDTendermint(query string) string {
	cmd := fmt.Sprintf("%s tendermint %s --home=%s", f.BkdBinary, query, f.BkdHome)
	success, stdout, stderr := executeWriteRetStdStreams(f.T, cmd)
	require.Empty(f.T, stderr)
	require.True(f.T, success)
	return strings.TrimSpace(stdout)
}

// ValidateGenesis runs blackd validate-genesis
func (f *Fixtures) ValidateGenesis() {
	cmd := fmt.Sprintf("%s validate-genesis --home=%s", f.BkdBinary, f.BkdHome)
	executeWriteCheckErr(f.T, cmd)
}

//___________________________________________________________________________________
// bkcli keys

// KeysDelete is bkcli keys delete
func (f *Fixtures) KeysDelete(name string, flags ...string) {
	cmd := fmt.Sprintf("%s keys delete --keyring-backend=test --home=%s %s", f.BkcliBinary,
		f.BkcliHome, name)
	executeWrite(f.T, addFlags(cmd, append(append(flags, "-y"), "-f")))
}

// KeysAdd is bkcli keys add
func (f *Fixtures) KeysAdd(name string, flags ...string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s %s", f.BkcliBinary,
		f.BkcliHome, name)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

// KeysAddRecover prepares bkcli keys add --recover
func (f *Fixtures) KeysAddRecover(name, mnemonic string, flags ...string) (exitSuccess bool, stdout, stderr string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s --recover %s",
		f.BkcliBinary, f.BkcliHome, name)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), mnemonic)
}

// KeysAddRecoverHDPath prepares bkcli keys add --recover --account --index
func (f *Fixtures) KeysAddRecoverHDPath(name, mnemonic string, account uint32, index uint32, flags ...string) {
	cmd := fmt.Sprintf("%s keys add --keyring-backend=test --home=%s --recover %s --account %d"+
		" --index %d", f.BkcliBinary, f.BkcliHome, name, account, index)
	executeWriteCheckErr(f.T, addFlags(cmd, flags), mnemonic)
}

// KeysShow is bkcli keys show
func (f *Fixtures) KeysShow(name string, flags ...string) keys.KeyOutput {
	cmd := fmt.Sprintf("%s keys show --keyring-backend=test --home=%s %s", f.BkcliBinary,
		f.BkcliHome, name)
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var ko keys.KeyOutput
	err := clientkeys.UnmarshalJSON([]byte(out), &ko)
	require.NoError(f.T, err)
	return ko
}

// KeyAddress returns the SDK account address from the key
func (f *Fixtures) KeyAddress(name string) sdk.AccAddress {
	ko := f.KeysShow(name)
	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(f.T, err)
	return accAddr
}

//___________________________________________________________________________________
// bkcli config

// CLIConfig is bkcli config
func (f *Fixtures) CLIConfig(key, value string, flags ...string) {
	cmd := fmt.Sprintf("%s config --home=%s %s %s", f.BkcliBinary, f.BkcliHome, key, value)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

//___________________________________________________________________________________
// bkcli tx send/sign/broadcast

// Status is bkcli status
func (f *Fixtures) Status(flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s status %s", f.BkcliBinary, f.Flags())
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxSend is bkcli tx send
func (f *Fixtures) TxSend(from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx send --keyring-backend=test %s %s %s %v", f.BkcliBinary, from,
		to, amount, f.Flags())
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxSign is bkcli tx sign
func (f *Fixtures) TxSign(signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign %v --keyring-backend=test --from=%s %v", f.BkcliBinary,
		f.Flags(), signer, fileName)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxBroadcast is bkcli tx broadcast
func (f *Fixtures) TxBroadcast(fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx broadcast %v %v", f.BkcliBinary, f.Flags(), fileName)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxEncode is bkcli tx encode
func (f *Fixtures) TxEncode(fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx encode %v %v", f.BkcliBinary, f.Flags(), fileName)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxMultisign is bkcli tx multisign
func (f *Fixtures) TxMultisign(fileName, name string, signaturesFiles []string,
	flags ...string,
) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx multisign --keyring-backend=test %v %s %s %s", f.BkcliBinary, f.Flags(),
		fileName, name, strings.Join(signaturesFiles, " "),
	)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags))
}

//___________________________________________________________________________________
// bkcli tx staking

// TxStakingCreateValidator is bkcli tx staking create-validator
func (f *Fixtures) TxStakingCreateValidator(from, consPubKey string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking create-validator %v --keyring-backend=test --from=%s"+
		" --pubkey=%s", f.BkcliBinary, f.Flags(), from, consPubKey)
	cmd += fmt.Sprintf(" --amount=%v --moniker=%v --commission-rate=%v", amount, from, "0.05")
	cmd += fmt.Sprintf(" --commission-max-rate=%v --commission-max-change-rate=%v", "0.20", "0.10")
	cmd += fmt.Sprintf(" --min-self-delegation=%v", "1")
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingUnbond is bkcli tx staking unbond
func (f *Fixtures) TxStakingUnbond(from, shares string, validator sdk.ValAddress, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx staking unbond --keyring-backend=test %s %v --from=%s %v",
		f.BkcliBinary, validator, shares, from, f.Flags())
	return executeWrite(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

//___________________________________________________________________________________
// bkcli tx gov

// TxGovSubmitProposal is bkcli tx gov submit-proposal
func (f *Fixtures) TxGovSubmitProposal(from, typ, title, description string, deposit sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov submit-proposal %v --keyring-backend=test --from=%s --type=%s",
		f.BkcliBinary, f.Flags(), from, typ)
	cmd += fmt.Sprintf(" --title=%s --description=%s --deposit=%s", title, description, deposit)
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovDeposit is bkcli tx gov deposit
func (f *Fixtures) TxGovDeposit(proposalID int, from string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov deposit %d %s --keyring-backend=test --from=%s %v",
		f.BkcliBinary, proposalID, amount, from, f.Flags())
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovVote is bkcli tx gov vote
func (f *Fixtures) TxGovVote(proposalID int, option gov.VoteOption, from string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx gov vote %d %s --keyring-backend=test --from=%s %v",
		f.BkcliBinary, proposalID, option, from, f.Flags())
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovSubmitParamChangeProposal executes a CLI parameter change proposal
// submission.
func (f *Fixtures) TxGovSubmitParamChangeProposal(
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (bool, string, string) {
	cmd := fmt.Sprintf(
		"%s tx gov submit-proposal param-change %s --keyring-backend=test --from=%s %v",
		f.BkcliBinary, proposalPath, from, f.Flags(),
	)

	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxGovSubmitCommunityPoolSpendProposal executes a CLI community pool spend proposal
// submission.
func (f *Fixtures) TxGovSubmitCommunityPoolSpendProposal(
	from, proposalPath string, deposit sdk.Coin, flags ...string,
) (bool, string, string) {
	cmd := fmt.Sprintf(
		"%s tx gov submit-proposal community-pool-spend %s --keyring-backend=test --from=%s %v",
		f.BkcliBinary, proposalPath, from, f.Flags(),
	)

	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

//___________________________________________________________________________________
// bkcli query account

// QueryAccount is bkcli query account
func (f *Fixtures) QueryAccount(address sdk.AccAddress, flags ...string) auth.BaseAccount {
	cmd := fmt.Sprintf("%s query account %s %v", f.BkcliBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(f.T, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	err = cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)
	return acc
}

//___________________________________________________________________________________
// bkcli query txs

// QueryTxs is bkcli query txs
func (f *Fixtures) QueryTxs(page, limit int, events ...string) *sdk.SearchTxsResult {
	cmd := fmt.Sprintf("%s query txs --page=%d --limit=%d --events='%s' %v", f.BkcliBinary, page, limit, queryEvents(events), f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var result sdk.SearchTxsResult

	err := f.cdc.UnmarshalJSON([]byte(out), &result)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return &result
}

// QueryTxsInvalid query txs with wrong parameters and compare expected error
func (f *Fixtures) QueryTxsInvalid(expectedErr error, page, limit int, events ...string) {
	cmd := fmt.Sprintf("%s query txs --page=%d --limit=%d --events='%s' %v", f.BkcliBinary, page, limit, queryEvents(events), f.Flags())
	_, err := tests.ExecuteT(f.T, cmd, "")
	require.EqualError(f.T, expectedErr, err)
}

//___________________________________________________________________________________
// bkcli query staking

// QueryStakingValidator is bkcli query staking validator
func (f *Fixtures) QueryStakingValidator(valAddr sdk.ValAddress, flags ...string) staking.Validator {
	cmd := fmt.Sprintf("%s query staking validator %s %v", f.BkcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var validator staking.Validator

	err := f.cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return validator
}

// QueryStakingUnbondingDelegationsFrom is bkcli query staking unbonding-delegations-from
func (f *Fixtures) QueryStakingUnbondingDelegationsFrom(valAddr sdk.ValAddress, flags ...string) []staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegations-from %s %v", f.BkcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var ubds []staking.UnbondingDelegation

	err := f.cdc.UnmarshalJSON([]byte(out), &ubds)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return ubds
}

// QueryStakingDelegationsTo is bkcli query staking delegations-to
func (f *Fixtures) QueryStakingDelegationsTo(valAddr sdk.ValAddress, flags ...string) []staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegations-to %s %v", f.BkcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var delegations []staking.Delegation

	err := f.cdc.UnmarshalJSON([]byte(out), &delegations)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return delegations
}

// QueryStakingPool is bkcli query staking pool
func (f *Fixtures) QueryStakingPool(flags ...string) staking.Pool {
	cmd := fmt.Sprintf("%s query staking pool %v", f.BkcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var pool staking.Pool

	err := f.cdc.UnmarshalJSON([]byte(out), &pool)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return pool
}

// QueryStakingParameters is bkcli query staking parameters
func (f *Fixtures) QueryStakingParameters(flags ...string) staking.Params {
	cmd := fmt.Sprintf("%s query staking params %v", f.BkcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var params staking.Params

	err := f.cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return params
}

//___________________________________________________________________________________
// bkcli query gov

// QueryGovParamDeposit is bkcli query gov param deposit
func (f *Fixtures) QueryGovParamDeposit() gov.DepositParams {
	cmd := fmt.Sprintf("%s query gov param deposit %s", f.BkcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var depositParam gov.DepositParams

	err := f.cdc.UnmarshalJSON([]byte(out), &depositParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return depositParam
}

// QueryGovParamVoting is bkcli query gov param voting
func (f *Fixtures) QueryGovParamVoting() gov.VotingParams {
	cmd := fmt.Sprintf("%s query gov param voting %s", f.BkcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var votingParam gov.VotingParams

	err := f.cdc.UnmarshalJSON([]byte(out), &votingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votingParam
}

// QueryGovParamTallying is bkcli query gov param tallying
func (f *Fixtures) QueryGovParamTallying() gov.TallyParams {
	cmd := fmt.Sprintf("%s query gov param tallying %s", f.BkcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, cmd, "")
	var tallyingParam gov.TallyParams

	err := f.cdc.UnmarshalJSON([]byte(out), &tallyingParam)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return tallyingParam
}

// QueryGovProposals is bkcli query gov proposals
func (f *Fixtures) QueryGovProposals(flags ...string) gov.Proposals {
	cmd := fmt.Sprintf("%s query gov proposals %v", f.BkcliBinary, f.Flags())
	stdout, stderr := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	if strings.Contains(stderr, "no matching proposals found") {
		return gov.Proposals{}
	}
	require.Empty(f.T, stderr)
	var out gov.Proposals

	err := f.cdc.UnmarshalJSON([]byte(stdout), &out)
	require.NoError(f.T, err)
	return out
}

// QueryGovProposal is bkcli query gov proposal
func (f *Fixtures) QueryGovProposal(proposalID int, flags ...string) gov.Proposal {
	cmd := fmt.Sprintf("%s query gov proposal %d %v", f.BkcliBinary, proposalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var proposal gov.Proposal

	err := f.cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return proposal
}

// QueryGovVote is bkcli query gov vote
func (f *Fixtures) QueryGovVote(proposalID int, voter sdk.AccAddress, flags ...string) gov.Vote {
	cmd := fmt.Sprintf("%s query gov vote %d %s %v", f.BkcliBinary, proposalID, voter, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var vote gov.Vote

	err := f.cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return vote
}

// QueryGovVotes is bkcli query gov votes
func (f *Fixtures) QueryGovVotes(proposalID int, flags ...string) []gov.Vote {
	cmd := fmt.Sprintf("%s query gov votes %d %v", f.BkcliBinary, proposalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var votes []gov.Vote

	err := f.cdc.UnmarshalJSON([]byte(out), &votes)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return votes
}

// QueryGovDeposit is bkcli query gov deposit
func (f *Fixtures) QueryGovDeposit(proposalID int, depositor sdk.AccAddress, flags ...string) gov.Deposit {
	cmd := fmt.Sprintf("%s query gov deposit %d %s %v", f.BkcliBinary, proposalID, depositor, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var deposit gov.Deposit

	err := f.cdc.UnmarshalJSON([]byte(out), &deposit)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposit
}

// QueryGovDeposits is bkcli query gov deposits
func (f *Fixtures) QueryGovDeposits(propsalID int, flags ...string) []gov.Deposit {
	cmd := fmt.Sprintf("%s query gov deposits %d %v", f.BkcliBinary, propsalID, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var deposits []gov.Deposit

	err := f.cdc.UnmarshalJSON([]byte(out), &deposits)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return deposits
}

//___________________________________________________________________________________
// query slashing

// QuerySigningInfo returns the signing info for a validator
func (f *Fixtures) QuerySigningInfo(val string) slashing.ValidatorSigningInfo {
	cmd := fmt.Sprintf("%s query slashing signing-info %s %s", f.BkcliBinary, val, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var sinfo slashing.ValidatorSigningInfo
	err := f.cdc.UnmarshalJSON([]byte(res), &sinfo)
	require.NoError(f.T, err)
	return sinfo
}

// QuerySlashingParams is bkcli query slashing params
func (f *Fixtures) QuerySlashingParams() slashing.Params {
	cmd := fmt.Sprintf("%s query slashing params %s", f.BkcliBinary, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var params slashing.Params
	err := f.cdc.UnmarshalJSON([]byte(res), &params)
	require.NoError(f.T, err)
	return params
}

//___________________________________________________________________________________
// query distribution

// QueryRewards returns the rewards of a delegator
func (f *Fixtures) QueryRewards(delAddr sdk.AccAddress, flags ...string) distribution.QueryDelegatorTotalRewardsResponse {
	cmd := fmt.Sprintf("%s query distribution rewards %s %s", f.BkcliBinary, delAddr, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var rewards distribution.QueryDelegatorTotalRewardsResponse
	err := f.cdc.UnmarshalJSON([]byte(res), &rewards)
	require.NoError(f.T, err)
	return rewards
}

//___________________________________________________________________________________
// query supply

// QueryTotalSupply returns the total supply of coins
func (f *Fixtures) QueryTotalSupply(flags ...string) (totalSupply sdk.Coins) {
	cmd := fmt.Sprintf("%s query supply total %s", f.BkcliBinary, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	err := f.cdc.UnmarshalJSON([]byte(res), &totalSupply)
	require.NoError(f.T, err)
	return totalSupply
}

// QueryTotalSupplyOf returns the total supply of a given coin denom
func (f *Fixtures) QueryTotalSupplyOf(denom string, flags ...string) sdkmath.Int {
	cmd := fmt.Sprintf("%s query supply total %s %s", f.BkcliBinary, denom, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var supplyOf sdkmath.Int
	err := f.cdc.UnmarshalJSON([]byte(res), &supplyOf)
	require.NoError(f.T, err)
	return supplyOf
}

//___________________________________________________________________________________
// executors

func executeWriteCheckErr(t *testing.T, cmdStr string, writes ...string) {
	require.True(t, executeWrite(t, cmdStr, writes...))
}

func executeWrite(t *testing.T, cmdStr string, writes ...string) (exitSuccess bool) {
	exitSuccess, _, _ = executeWriteRetStdStreams(t, cmdStr, writes...)
	return
}

func executeWriteRetStdStreams(t *testing.T, cmdStr string, writes ...string) (bool, string, string) {
	proc := tests.GoExecuteT(t, cmdStr)

	// Enables use of interactive commands
	for _, write := range writes {
		_, err := proc.StdinPipe.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}

	// Read both stdout and stderr from the process
	stdout, stderr, err := proc.ReadAll()
	if err != nil {
		fmt.Println("Err on proc.ReadAll()", err, cmdStr)
	}

	// Log output.
	if len(stdout) > 0 {
		t.Log("Stdout:", string(stdout))
	}
	if len(stderr) > 0 {
		t.Log("Stderr:", string(stderr))
	}

	// Wait for process to exit
	proc.Wait()

	// Return succes, stdout, stderr
	return proc.ExitState.Success(), string(stdout), string(stderr)
}

//___________________________________________________________________________________
// utils

func addFlags(cmd string, flags []string) string {
	for _, f := range flags {
		cmd += " " + f
	}
	return strings.TrimSpace(cmd)
}

func queryEvents(events []string) (out string) {
	for _, event := range events {
		out += event + "&"
	}
	return strings.TrimSuffix(out, "&")
}

// Write the given string to a new temporary file
func WriteToNewTempFile(t *testing.T, s string) *os.File {
	fp, err := ioutil.TempFile(os.TempDir(), "cosmos_cli_test_")
	require.Nil(t, err)
	_, err = fp.WriteString(s)
	require.Nil(t, err)
	return fp
}

//nolint:deadcode,unused
func (f *Fixtures) marshalStdTx(t *testing.T, stdTx auth.StdTx) []byte {
	bz, err := f.cdc.MarshalBinaryBare(stdTx)
	require.NoError(t, err)
	return bz
}

//nolint:deadcode,unused
func (f *Fixtures) unmarshalStdTx(t *testing.T, s string) (stdTx auth.StdTx) {
	require.Nil(t, f.cdc.UnmarshalJSON([]byte(s), &stdTx))
	return
}
