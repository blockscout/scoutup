package config

import (
	"encoding/json"
	"fmt"

	"github.com/blockscout/scoutup/utils"
)

type InstanceConfig struct {
	DockerRepo        string
	DockerTag         string
	FrontendDockerTag string
	FrontendPort      uint64
	BackendPort       uint64
	PostgresPort      uint64
}

type BlockscoutConfig struct {
	*ChainConfig
	*InstanceConfig
	// Maps (other than current) l2 chain ids to corresponding instance configs
	OtherL2InstanceConfigs map[uint64]*InstanceConfig
}

func (b *BlockscoutConfig) DockerComposeEnvs() []string {
	return []string{
		fmt.Sprintf("DOCKER_REPO=%s", b.DockerRepo),
		fmt.Sprintf("DOCKER_TAG=%s", b.DockerTag),
		fmt.Sprintf("FRONTEND_DOCKER_TAG=%s", b.FrontendDockerTag),
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

		envs["CHAIN_TYPE"] = "optimism"

		// interop image related envs
		envs["INDEXER_OPTIMISM_L2_INTEROP_START_BLOCK"] = "0"
		envs["INDEXER_OPTIMISM_INTEROP_PRIVATE_KEY"] = "0x5721810206e5e84fd05f9f0b9aa2d7544a3ea29674b24028d7a6a60d803a33a3"
		envs["INDEXER_OPTIMISM_CHAINSCOUT_FALLBACK_MAP"] = b.buildChainscoutFallbackMapValue()
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

		// interop image related envs
		envs["NEXT_PUBLIC_INTEROP_ENABLED"] = "true"
	}

	return envs
}

func (b *BlockscoutConfig) buildChainscoutFallbackMapValue() string {
	chainscoutFallbackMap := make(map[string]map[string]string)
	for chainID, config := range b.OtherL2InstanceConfigs {
		chainFallbackMap := make(map[string]string)
		chainFallbackMap["api"] = fmt.Sprintf("http://host.docker.internal:%v", config.BackendPort)
		chainFallbackMap["ui"] = fmt.Sprintf("http://127.0.0.1:%v", config.FrontendPort)
		chainscoutFallbackMap[fmt.Sprintf("%v", chainID)] = chainFallbackMap
	}
	jsonBytes, _ := json.Marshal(chainscoutFallbackMap)
	return string(jsonBytes)
}
