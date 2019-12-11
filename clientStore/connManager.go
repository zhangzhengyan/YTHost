package clientStore

import (
	"context"
	"fmt"
	"github.com/graydream/YTHost/client"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multiaddr"
)

type ClientStore struct {
	psmap   map[peer.ID]*client.YTHostClient
	connect func(ctx context.Context, id peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error)
}

// Get 获取一个客户端，如果没有，建立新的客户端连接
func (cs *ClientStore) Get(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx done")
	default:
		return cs.get(ctx, pid, mas)
	}
}

func (cs *ClientStore) get(ctx context.Context, pid peer.ID, mas []multiaddr.Multiaddr) (*client.YTHostClient, error) {
	// 尝试次数
	var tryCount int
	const max_try_count = 5

	// 取已存在clt
start:
	// 如果达到最大尝试次数就返回错误
	if tryCount++; tryCount > max_try_count {
		return nil, fmt.Errorf("Maximum attempts %d ", max_try_count)
	}

	c, ok := cs.psmap[pid]
	// 如果不存在创建新的clt
	if !ok || c.IsClosed() {
		if clt, err := cs.connect(ctx, pid, mas); err != nil {
			return nil, err
		} else {
			cs.psmap[pid] = clt
			// 创建clt完成后返回到开始
			goto start
		}
	} else {
		// 如果已存在clt无法ping通,删除记录重新创建
		if !c.Ping(ctx) {
			delete(cs.psmap, pid)
			goto start
		}
	}
	return c, nil
}

func (cs *ClientStore) GetByAddrString(ctx context.Context, id string, addrs []string) (*client.YTHostClient, error) {
	buf, _ := base58.Decode(id)
	pid, err := peer.IDFromBytes(buf)
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
