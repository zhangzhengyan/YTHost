package option

import (
	"github.com/graydream/YTHost/config"
	"github.com/multiformats/go-multiaddr"
)

type Option func (cfg *config.Config)

func ListenAddr(ma multiaddr.Multiaddr) Option {
	return func(cfg *config.Config) {
		cfg.ListenAddr = ma
	}
}