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
	RegisterHandler(id service.MsgId, handlerFunc service.Handler)
	RegisterGlobalMsgHandler(handlerFunc service.Handler)
}
