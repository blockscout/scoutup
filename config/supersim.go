package config

import (
	"fmt"

	ssconfig "github.com/ethereum-optimism/supersim/config"
	"github.com/ethereum/go-ethereum/rpc"
)

func PrepareSupersimConfig(admingRPCUrl string) (*NetworkConfig, error) {
	sc, err := fetchSupersimConfig(admingRPCUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch supersim config: %w", err)
	}

	l1Config := &ChainConfig{
		Name:        sc.L1Config.Name,
		RPCUrl:      fmt.Sprintf("http://host.docker.internal:%d", sc.L1Config.Port),
		ChainID:     sc.L1Config.ChainID,
		GenesisJSON: sc.L1Config.GenesisJSON,
	}
	if sc.L1Config.ForkConfig != nil && sc.L1Config.ForkConfig.BlockNumber > 0 {
		l1Config.FirstBlock = sc.L1Config.ForkConfig.BlockNumber
	}

	networkConfig := &NetworkConfig{
		Chains: []*ChainConfig{
			l1Config,
		},
	}

	for i, chain := range sc.L2Configs {
		port := sc.L2StartingPort + uint64(i)
		if chain.Port > 0 {
			port = chain.Port
		}

		l2Config := &ChainConfig{
			Name:    chain.Name,
			RPCUrl:  fmt.Sprintf("http://host.docker.internal:%d", port),
			ChainID: chain.ChainID,
			OPConfig: &OPConfig{
				L1RPCUrl:               fmt.Sprintf("http://host.docker.internal:%d", sc.L1Config.Port),
				L1SystemConfigContract: chain.L2Config.L1Addresses.SystemConfigProxy.String(),
			},
			GenesisJSON: chain.GenesisJSON,
		}

		if chain.ForkConfig != nil && chain.ForkConfig.BlockNumber > 0 {
			l2Config.FirstBlock = chain.ForkConfig.BlockNumber
		}

		networkConfig.Chains = append(networkConfig.Chains, l2Config)
	}

	return networkConfig, nil
}

func fetchSupersimConfig(admingRPCUrl string) (*ssconfig.NetworkConfig, error) {
	client, err := rpc.DialHTTP(admingRPCUrl)
	if err != nil {
		return nil, err
	}

	var config *ssconfig.NetworkConfig
	err = client.Call(&config, "admin_getConfig")
	if err != nil {
		return nil, err
	}

	return config, nil
}
