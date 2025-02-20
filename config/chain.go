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
	OPConfig    *OPConfig
	GenesisJSON []byte
}

func (n *ChainConfig) dockerRepo() string {
	if n.OPConfig != nil {
		return "blockscout-optimism"
	}
	return "blockscout"
}
