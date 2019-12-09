package clientStore

import (
	"context"
	"fmt"
	"github.com/eoscanada/eos-go/btcsuite/btcutil/base58"
	"github.com/graydream/YTHost/client"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type ClientStore struct {
	psmap   map[peer.ID]*client.YTHostClient
	connect func(ctx context.Context, id peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error)
}

// Get 获取一个客户端，如果没有，建立新的客户端连接
func (cs *ClientStore) Get(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error) {
	c, ok := cs.psmap[pid]
	if !ok || c.IsClosed() {
		if clt, err := cs.connect(ctx, pid, mas); err != nil {
			return nil, err
		} else {
			cs.psmap[pid] = clt
		}
	}
	return cs.psmap[pid], nil
}

func (cs *ClientStore) GetByAddrString(ctx context.Context, id string, addrs []string) (*client.YTHostClient, error) {
	pid, err := peer.IDFromBytes(base58.Decode(id))
	if err != nil {
		return nil, err
	}

	var mas = make([]multiaddr.Multiaddr, len(addrs))
	for k, v := range addrs {
		ma, err := multiaddr.NewMultiaddr(v)
		if err != nil {
			continue
		}
		mas[k] = ma
	}
	c, ok := cs.psmap[pid]
	if !ok || c.IsClosed() {
		if clt, err := cs.connect(ctx, pid, mas); err != nil {
			return nil, err
		} else {
			cs.psmap[pid] = clt
		}
	}
	return cs.psmap[pid], nil
}

// Close 关闭一个客户端
func (cs *ClientStore) Close(pid peer.ID) error {
	clt, ok := cs.psmap[pid]
	if !ok {
		return fmt.Errorf("no find client ID is %s", pid.Pretty())
	}
	delete(cs.psmap, pid)
	return clt.Close()
}

func (cs *ClientStore) GetClient(pid peer.ID) (*client.YTHostClient, bool) {
	clt, ok := cs.psmap[pid]
	return clt, ok
}

// Len 返回当前连接数
func (cs *ClientStore) Len() int {
	return len(cs.psmap)
}

func NewClientStore(connFunc func(ctx context.Context, id peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error)) *ClientStore {
	return &ClientStore{
		make(map[peer.ID]*client.YTHostClient),
		connFunc,
	}
}
