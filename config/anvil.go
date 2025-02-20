package config

func PrepareDefaultAnvilConfig() *NetworkConfig {
	return &NetworkConfig{
		Chains: []*ChainConfig{
			{
				Name:       "Local Anvil",
				RPCUrl:     "http://host.docker.internal:8545",
				FirstBlock: 0,
				ChainID:    900,
			},
		},
	}
}
