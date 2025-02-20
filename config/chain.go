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
