package hostInterface

import (
	"context"
	"github.com/graydream/YTHost/client"
	"github.com/graydream/YTHost/config"
	"github.com/graydream/YTHost/service"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"net/rpc"
)

type Host interface {
	Accept()
	Addrs() []multiaddr.Multiaddr
	Server() *rpc.Server
	Config() *config.Config
	Connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error)
	RegisterHandler(id int32, handlerFunc service.Handler) error
	RegisterGlobalMsgHandler(handlerFunc service.Handler)
	RemoveMsgHandler(id int32)
	RemoveGlobalMsgHandler()
	ConnectAddrStrings(ctx context.Context, id string, addrs []string) (*client.YTHostClient, error)
}
