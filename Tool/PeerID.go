package Tool

import (
	"crypto/rand"
	"github.com/libp2p/go-libp2p-core/crypto"
	b58 "github.com/mr-tron/base58/base58"
	mh "github.com/multiformats/go-multihash"
	"io"
)

type ID string

func GetPeerId() (string, error) {
	_, pubKey, err := GetPrvPubKey()
	if err != nil {
		return "", err
	}
	ID, err := GetPeerIdFromPubKey(pubKey)
	if err != nil {
		return "", err
	}
	return ID.Pretty(), nil
}
// 获取公私钥对
func GetPrvPubKey()  (crypto.PrivKey, crypto.PubKey, error) {
	var r io.Reader
	r = rand.Reader
	prvKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	return prvKey, pubKey, err
}

// 根据公钥生成peerId
func GetPeerIdFromPubKey(key crypto.PubKey) (ID, error) {
	b, err := key.Bytes()
	if err != nil {
		return "", err
	}
	var alg uint64 = mh.SHA2_256
	//if AdvancedEnableInlining && len(b) <= maxInlineKeyLength {
	//	alg = mh.ID
	//}
	hash, _ := mh.Sum(b, alg, -1)
	return ID(hash), nil
}

func (id ID) Pretty() string {
	return IDB58Encode(id)
}

func (id ID) String() string {
	return id.Pretty()
}

func IDB58Encode(id ID) string {
	return b58.Encode([]byte(id))
}