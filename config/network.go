package config

import "fmt"

type NetworkConfig struct {
	Chains               []*ChainConfig
	StartingFrontendPort uint64
	StartingBackendPort  uint64
	StartingPostgresPort uint64
}

func (n *NetworkConfig) PrepareBlockscoutConfigs() []*BlockscoutConfig {
	frontendPort := n.StartingFrontendPort
	backendPort := n.StartingBackendPort
	postgresPort := n.StartingPostgresPort

	configs := []*BlockscoutConfig{}
	for _, chain := range n.Chains {
		config := &BlockscoutConfig{
			ChainConfig: chain,
			InstanceConfig: &InstanceConfig{
				FrontendPort: frontendPort,
				BackendPort:  backendPort,
				PostgresPort: postgresPort,
				DockerRepo:   chain.dockerRepo(),
			},
		}

		if config.OPConfig != nil {
			// TODO: refactor this later
			for _, bs := range configs {
				if bs.RPCUrl == config.OPConfig.L1RPCUrl {
					config.OPConfig.L1BlockscoutURL = fmt.Sprintf("http://host.docker.internal:%v", bs.FrontendPort)
					break
				}
			}
		}

		configs = append(configs, config)
		frontendPort++
		backendPort++
		postgresPort++
	}
	return configs
}
