package config

import (
	"fmt"

	"github.com/blockscout/scoutup/utils"
)

type InstanceConfig struct {
	DockerRepo   string
	FrontendPort uint64
	BackendPort  uint64
	PostgresPort uint64
}

type BlockscoutConfig struct {
	*ChainConfig
	*InstanceConfig
}

func (b *BlockscoutConfig) DockerComposeEnvs() []string {
	return []string{
		fmt.Sprintf("DOCKER_REPO=%s", b.DockerRepo),
		fmt.Sprintf("FRONTEND_PORT=%d", b.FrontendPort),
		fmt.Sprintf("BACKEND_PORT=%d", b.BackendPort),
		fmt.Sprintf("POSTGRES_PORT=%d", b.PostgresPort),
		fmt.Sprintf("DB_CONTAINER_NAME=%s", utils.NameToContainerName("db", b.Name)),
		fmt.Sprintf("BACKEND_CONTAINER_NAME=%s", utils.NameToContainerName("backend", b.Name)),
		fmt.Sprintf("FRONTEND_CONTAINER_NAME=%s", utils.NameToContainerName("frontend", b.Name)),
	}
}

func (b *BlockscoutConfig) BackendEnvs() map[string]string {
	envs := make(map[string]string)
	envs["ETHEREUM_JSONRPC_HTTP_URL"] = b.RPCUrl
	envs["ETHEREUM_JSONRPC_TRACE_URL"] = b.RPCUrl
	envs["SUBNETWORK"] = b.Name
	envs["FIRST_BLOCK"] = fmt.Sprintf("%d", b.FirstBlock)
	envs["DATABASE_URL"] = fmt.Sprintf(
		"postgresql://blockscout:ceWb1MeLBEeOIfk65gU8EjF8@host.docker.internal:%v/blockscout", b.PostgresPort)
	envs["CHAIN_SPEC_PATH"] = "/app/genesis.json"
	envs["CHAIN_SPEC_PROCESSING_DELAY"] = "0s"

	if b.OPConfig != nil {
		envs["INDEXER_OPTIMISM_L1_RPC"] = b.OPConfig.L1RPCUrl
		envs["INDEXER_OPTIMISM_L1_SYSTEM_CONFIG_CONTRACT"] = b.OPConfig.L1SystemConfigContract
		envs["INDEXER_OPTIMISM_L2_BATCH_GENESIS_BLOCK_NUMBER"] = "0"
		envs["INDEXER_OPTIMISM_L2_HOLOCENE_TIMESTAMP"] = "0"
	}
	return envs
}

func (b *BlockscoutConfig) FrontendEnvs() map[string]string {
	envs := make(map[string]string)
	envs["NEXT_PUBLIC_API_PORT"] = fmt.Sprintf("%d", b.BackendPort)
	envs["NEXT_PUBLIC_NETWORK_NAME"] = b.Name
	envs["NEXT_PUBLIC_NETWORK_SHORT_NAME"] = b.Name

	if b.OPConfig != nil {
		envs["NEXT_PUBLIC_ROLLUP_TYPE"] = "optimistic"
		envs["NEXT_PUBLIC_ROLLUP_L1_BASE_URL"] = b.OPConfig.L1BlockscoutURL
		// TODO: what is the correct value here?
		envs["NEXT_PUBLIC_ROLLUP_L2_WITHDRAWAL_URL"] = "https://app.optimism.io/bridge/withdraw"
	}

	return envs
}
