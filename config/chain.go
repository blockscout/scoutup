package config

type OPConfig struct {
	L1RPCUrl               string
	L1SystemConfigContract string
	L1BlockscoutURL        string
}

type ChainConfig struct {
	Name        string
	RPCUrl      string
	FirstBlock  uint64
	ChainID     uint64
	GenesisJSON []byte
	OPConfig    *OPConfig
}

func (n *ChainConfig) dockerRepo() string {
	if n.OPConfig != nil {
		return "blockscout-optimism"
	}
	return "blockscout"
}

func (n *ChainConfig) dockerTag() string {
	if n.OPConfig != nil {
		return "7.0.0-postrelease-bac46e76"
	}
	return "7.0.0"
}

func (n *ChainConfig) frontendDockerTag() string {
	if n.OPConfig != nil {
		return "interop"
	}
	return "v1.37.4"
}
