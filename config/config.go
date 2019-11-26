package config

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type Config struct {
	ListenAddr ma.Multiaddr
	Privkey    crypto.PrivKey
	ID         peer.ID
}

func NewConfig() *Config {
	cfg := new(Config)
	return newConfig(cfg)
}
func newConfig(cfg *Config) *Config {
	maddr, _ := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/9001")
	cfg.ListenAddr = maddr
	return cfg
}