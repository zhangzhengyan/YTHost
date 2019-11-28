package service

import (
	"github.com/libp2p/go-libp2p-core/peer"
)

type AddrService struct {
	Info peer.AddrInfo
}

type PeerInfo struct {
	ID    peer.ID
	Addrs []string
}

func (as *AddrService) RemotePeerInfo(req string, res *PeerInfo) error {
	res.ID = as.Info.ID
	//res.Addrs = as.Info.Addrs
	for _, addr := range as.Info.Addrs {
		res.Addrs = append(res.Addrs, addr.String())
	}
	return nil
}
