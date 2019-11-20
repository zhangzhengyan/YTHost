package host

import(
	ma "github.com/multiformats/go-multiaddr"
)

type Config struct {
	ListenAddr ma.Multiaddr
}

func NewConfig() *Config{
	cfg:=new(Config)
	newConfig(cfg)
	maddr,_ := ma.NewMultiaddr("/ip4/0.0.0.0/tcp/9001")
	cfg.ListenAddr = maddr
	return cfg
}
func newConfig(cfg *Config){
}
