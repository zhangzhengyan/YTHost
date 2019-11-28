package service

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
)

type AddrService struct {
	Info peer.AddrInfo
}

func (as *AddrService) RemotePeerInfo(req string, res *peer.AddrInfo) error {
	res.ID = as.Info.ID
	//res.Addrs = as.Info.Addrs
	for _, addr := range as.Info.Addrs {
		fmt.Println(addr.String())
	}
	return nil
}
