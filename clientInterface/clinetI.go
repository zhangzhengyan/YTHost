package clientInterface

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
)

type YTHClient interface{
	SendMsg(ctx context.Context, i int32, bytes []byte) ([]byte, error)
	IsClosed() bool
	Ping(ctx context.Context) bool
	Close() error
	LocalPeer() peer.AddrInfo
	SendMsgClose(ctx context.Context, id int32, data []byte) ([]byte, error)
	RemotePeer() peer.AddrInfo
	GetRemotePeerID() peer.ID
}
