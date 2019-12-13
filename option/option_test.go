package option_test

import (
	ra "crypto/rand"
	"github.com/graydream/YTHost"
	"github.com/graydream/YTHost/option"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
	"testing"
)

func TestOption(t *testing.T) {
	var localMa = "/ip4/0.0.0.0/tcp/9003"

	pk , _, _ := ic.GenerateSecp256k1Key(ra.Reader)		//option Identity cover
	ma, _ := multiaddr.NewMultiaddr(localMa)			// option listenaddr cover
	if hst, err := host.NewHost(option.ListenAddr(ma), option.Identity(pk)); err != nil {
		t.Fatal(err)
	} else {
		maddrs := hst.Addrs()
		for _, ma := range maddrs {
			t.Log(ma)
		}
		t.Log(hst.Config().ID)
	}

}
