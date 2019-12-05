package hostInterface

import (
	"context"
	"github.com/graydream/YTHost/client"
	"github.com/graydream/YTHost/clientStore"
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
	RemoveHandler(id int32)
	RemoveGlobalHandler()
	ConnectAddrStrings(ctx context.Context, id string, addrs []string) (*client.YTHostClient, error)
	ClientStore() *clientStore.ClientStore
	SendMsg(ctx context.Context, pid peer.ID, mid int32, msg []byte) ([]byte, error)
}
