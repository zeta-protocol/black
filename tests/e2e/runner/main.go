package runner

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type Config struct {
	BlackConfigTemplate string

	ImageTag   string
	IncludeIBC bool

	EnableAutomatedUpgrade  bool
	BlackUpgradeName         string
	BlackUpgradeHeight       int64
	BlackUpgradeBaseImageTag string

	SkipShutdown bool
}

// NodeRunner is responsible for starting and managing docker containers to run a node.
type NodeRunner interface {
	StartChains() Chains
	Shutdown()
}

// BlackNodeRunner manages and runs a single Black node.
type BlackNodeRunner struct {
	config    Config
	blackChain *ChainDetails
}

var _ NodeRunner = &BlackNodeRunner{}

func NewBlackNode(config Config) *BlackNodeRunner {
	return &BlackNodeRunner{
		config: config,
	}
}

func (k *BlackNodeRunner) StartChains() Chains {
	installKvtoolCmd := exec.Command("./scripts/install-kvtool.sh")
	installKvtoolCmd.Stdout = os.Stdout
	installKvtoolCmd.Stderr = os.Stderr
	if err := installKvtoolCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to install kvtool: %s", err.Error()))
	}

	log.Println("starting black node")
	kvtoolArgs := []string{"testnet", "bootstrap", "--black.configTemplate", k.config.BlackConfigTemplate}
	if k.config.IncludeIBC {
		kvtoolArgs = append(kvtoolArgs, "--ibc")
	}
	if k.config.EnableAutomatedUpgrade {
		kvtoolArgs = append(kvtoolArgs,
			"--upgrade-name", k.config.BlackUpgradeName,
			"--upgrade-height", fmt.Sprint(k.config.BlackUpgradeHeight),
			"--upgrade-base-image-tag", k.config.BlackUpgradeBaseImageTag,
		)
	}
	startBlackCmd := exec.Command("kvtool", kvtoolArgs...)
	startBlackCmd.Env = os.Environ()
	startBlackCmd.Env = append(startBlackCmd.Env, fmt.Sprintf("BLACK_TAG=%s", k.config.ImageTag))
	startBlackCmd.Stdout = os.Stdout
	startBlackCmd.Stderr = os.Stderr
	log.Println(startBlackCmd.String())
	if err := startBlackCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to start black: %s", err.Error()))
	}

	k.blackChain = &blackChain

	err := k.waitForChainStart()
	if err != nil {
		k.Shutdown()
		panic(err)
	}
	log.Println("black is started!")

	chains := NewChains()
	chains.Register("black", k.blackChain)
	if k.config.IncludeIBC {
		chains.Register("ibc", &ibcChain)
	}
	return chains
}

func (k *BlackNodeRunner) Shutdown() {
	if k.config.SkipShutdown {
		log.Printf("would shut down but SkipShutdown is true")
		return
	}
	log.Println("shutting down black node")
	shutdownBlackCmd := exec.Command("kvtool", "testnet", "down")
	shutdownBlackCmd.Stdout = os.Stdout
	shutdownBlackCmd.Stderr = os.Stderr
	if err := shutdownBlackCmd.Run(); err != nil {
		panic(fmt.Sprintf("failed to shutdown kvtool: %s", err.Error()))
	}
}

func (k *BlackNodeRunner) waitForChainStart() error {
	// exponential backoff on trying to ping the node, timeout after 30 seconds
	b := backoff.NewExponentialBackOff()
	b.MaxInterval = 5 * time.Second
	b.MaxElapsedTime = 30 * time.Second
	if err := backoff.Retry(k.pingBlack, b); err != nil {
		return fmt.Errorf("failed to start & connect to chain: %s", err)
	}
	b.Reset()
	// the evm takes a bit longer to start up. wait for it to start as well.
	if err := backoff.Retry(k.pingEvm, b); err != nil {
		return fmt.Errorf("failed to start & connect to chain: %s", err)
	}
	return nil
}

func (k *BlackNodeRunner) pingBlack() error {
	log.Println("pinging black chain...")
	url := fmt.Sprintf("http://localhost:%s/status", k.blackChain.RpcPort)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("ping to status failed: %d", res.StatusCode)
	}
	log.Println("successfully started Black!")
	return nil
}

func (k *BlackNodeRunner) pingEvm() error {
	log.Println("pinging evm...")
	url := fmt.Sprintf("http://localhost:%s", k.blackChain.EvmPort)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	// when running, it should respond 405 to a GET request
	if res.StatusCode != 405 {
		return fmt.Errorf("ping to evm failed: %d", res.StatusCode)
	}
	log.Println("successfully pinged EVM!")
	return nil
}
