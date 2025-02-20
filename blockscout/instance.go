package blockscout

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/blockscout/scoutup/config"
	"github.com/blockscout/scoutup/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

type Instance struct {
	config    *config.BlockscoutConfig
	log       log.Logger
	workspace string

	cmd *exec.Cmd

	resourceCtx    context.Context
	resourceCancel context.CancelFunc
	closeApp       context.CancelCauseFunc

	stopped   atomic.Bool
	stoppedCh chan struct{}
}

func NewInstance(log log.Logger, closeApp context.CancelCauseFunc, config *config.BlockscoutConfig, globalWorkspace string) (*Instance, error) {
	workspace, err := createInstanceWorkspace(globalWorkspace, config.GenesisJSON)
	if err != nil {
		return nil, err
	}
	resCtx, resCancel := context.WithCancel(context.Background())
	return &Instance{
		config:         config,
		log:            log,
		workspace:      workspace,
		resourceCtx:    resCtx,
		resourceCancel: resCancel,
		closeApp:       closeApp,
		stoppedCh:      make(chan struct{}, 1),
	}, nil
}

func (i *Instance) Start(ctx context.Context) error {
	i.log.Info("Starting Blockscout", "chain", i.config.Name)

	err := i.configureBlockscout()
	if err != nil {
		return err
	}

	err = i.runDockerCompose(ctx)
	if err != nil {
		return err
	}

	go i.verifyL2InteropContracts()

	return nil
}

func (i *Instance) Stop(_ context.Context) error {
	i.log.Info("Stopping Blockscout", "chain", i.config.Name)
	if i.stopped.Load() {
		return errors.New("already stopped")
	}
	if !i.stopped.CompareAndSwap(false, true) {
		return nil // someone else stopped
	}

	i.resourceCancel()
	<-i.stoppedCh
	return nil
}

func (i *Instance) ConfigAsString() string {
	// TODO: prettify this
	var b strings.Builder
	fmt.Fprintf(&b, "* Chain: %v\n", i.config.Name)
	fmt.Fprintf(&b, "         Frontend:  http://127.0.0.1:%v\n", i.config.FrontendPort)
	fmt.Fprintf(&b, "         Backend:   http://127.0.0.1:%v\n", i.config.BackendPort)
	fmt.Fprintf(&b, "         DB:        http://127.0.0.1:%v\n", i.config.PostgresPort)
	fmt.Fprintf(&b, "         Workspace: %v\n", i.workspace)
	fmt.Fprintf(&b, "         Logs:	    %v\n", path.Join(i.workspace, "logs"))
	fmt.Fprintf(&b, "         First block: %v\n", i.config.FirstBlock)
	fmt.Fprintf(&b, "         RPC: %v\n", i.config.RPCUrl)
	fmt.Fprintf(&b, "         Chain ID: %v\n", i.config.ChainID)

	if i.config.OPConfig != nil {
		fmt.Fprintf(&b, "         Optimism L1 RPC: %v\n", i.config.OPConfig.L1RPCUrl)
		fmt.Fprintf(&b, "         Optimism L1 System Config Contract: %v\n", i.config.OPConfig.L1SystemConfigContract)
	}
	return b.String()
}

func (i *Instance) configureBlockscout() error {
	utils.PatchDotEnv(path.Join(i.workspace, "common-blockscout.env"), i.config.BackendEnvs())
	utils.PatchDotEnv(path.Join(i.workspace, "common-frontend.env"), i.config.FrontendEnvs())
	return nil
}

func (i *Instance) runDockerCompose(ctx context.Context) error {
	i.cmd = exec.CommandContext(i.resourceCtx, "docker", "compose", "up")
	i.cmd.Env = append(os.Environ(), i.config.DockerComposeEnvs()...)
	i.cmd.Cancel = func() error {
		return i.cmd.Process.Signal(syscall.SIGTERM)
	}
	i.cmd.Dir = i.workspace
	go func() {
		<-ctx.Done()
		i.resourceCancel()
	}()

	stdout, err := i.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	logFile, err := os.Create(path.Join(i.workspace, "logs"))
	if err != nil {
		return err
	}

	go func() {
		if _, err := io.Copy(logFile, stdout); err != nil {
			i.log.Warn("err piping stderr to log file", "err", err)
		}
	}()

	stderr, err := i.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go func() {
		if _, err := io.Copy(logFile, stderr); err != nil {
			i.log.Warn("err piping stderr to log file", "err", err)
		}
	}()

	if err := i.cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := i.cmd.Wait(); err != nil {
			if err.Error() != "exit status 130" {
				i.log.Error("Blockscout terminated with an error", "error", err)
			}
		} else {
			i.log.Info("Blockscout terminated")
		}

		err := cleanupInstanceWorkspace(i.workspace)
		if err != nil {
			i.log.Error("Failed to cleanup workspace", "error", err)
		}

		// If it stops, signal that the entire app should be closed
		i.closeApp(nil)
		i.stoppedCh <- struct{}{}
	}()

	return nil
}

func (i *Instance) verifyL2InteropContracts() {
	if i.config.OPConfig == nil {
		// the instance corresponds to L1 chain
		return
	}

	interopProxies := map[string]common.Address{
		"CrossL2Inbox":               common.HexToAddress("0x4200000000000000000000000000000000000022"),
		"L2ToL2CrossDomainMessenger": common.HexToAddress("0x4200000000000000000000000000000000000023"),
		"SuperchainWETH":             common.HexToAddress("0x4200000000000000000000000000000000000024"),
		"SuperchainTokenBridge":      common.HexToAddress("0x4200000000000000000000000000000000000028"),
	}

	backendURL := fmt.Sprintf("http://127.0.0.1:%d", i.config.BackendPort)

	// Wait for the backend instance to start
	for !isHealthy(backendURL) {
		time.Sleep(1 * time.Second)
	}

	i.log.Info("Verifying predeployed interop contracts", "chain", i.config.Name)

	interopImplementations := map[string][]common.Address{}
	for name, proxy := range interopProxies {
		implementations, err := retrieveProxyImplementationAddresses(backendURL, proxy)
		if err != nil {
			i.log.Error(
				"Failed to retrieve proxy implementation address",
				"chain", i.config.Name,
				"name", name,
				"err", err,
			)
			return
		}
		interopImplementations[name] = implementations
	}

	for name, implementations := range interopImplementations {
		for _, implementation := range implementations {
			// Should activate on-demand fetcher which retrieves sources from eth-bytecode-db
			_, err := utils.MakeGetRequest(getSmartContractUrl(backendURL, implementation))
			if err != nil {
				i.log.Error(
					"Failed to verify interop contract",
					"chain", i.config.Name,
					"name", name,
					"address", implementation,
					"err", err,
				)
			}
		}
	}
}
