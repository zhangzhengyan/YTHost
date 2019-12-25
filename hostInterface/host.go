package hostInterface

import (
	"context"
	ci "github.com/yottachain/YTHost/clientInterface"
	"github.com/yottachain/YTHost/clientStore"
	"github.com/yottachain/YTHost/config"
	"github.com/yottachain/YTHost/service"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"net/rpc"
	//"github.com/yottachain/YTHost/rpc"
)

type Host interface {
	Accept()
	Addrs() []multiaddr.Multiaddr
	Server() *rpc.Server
	Config() *config.Config
	Connect(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (ci.YTHClient, error)
	RegisterHandler(id int32, handlerFunc service.Handler) error
	RegisterGlobalMsgHandler(handlerFunc service.Handler)
	RemoveHandler(id int32)
	RemoveGlobalHandler()
	ConnectAddrStrings(ctx context.Context, id string, addrs []string) (ci.YTHClient, error)
	ClientStore() *clientStore.ClientStore
	SendMsg(ctx context.Context, pid peer.ID, mid int32, msg []byte) ([]byte, error)
}
