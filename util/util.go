package util

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"strings"
)

func MaddrRelayParse(smaddr string) (saddr string, relayID peer.ID, err error){
	strSplics := strings.Split(smaddr, "/")
	saddr = "/"
	lenth := len(strSplics)
	indx := 0
	for i, v := range strSplics {
		indx = i
		if "" != v  {
			if "p2p" == v {
				break
			}
			//fmt.Println(strSplics[i])
			saddr = saddr + v + "/"
		}
	}
	if indx + 1 != lenth {
		relayID, err = peer.Decode(strSplics[indx + 1])
	}else{
		relayID = ""
	}

	return
}

func MaddrToStringList( mas []multiaddr.Multiaddr ) ([]string){
	smaddrs := make([]string, len(mas))
	for k, addr := range mas {
		maddr := addr.String()
		smaddrs[k] = maddr
	}
	return smaddrs
}

func GetRealyId(mas []multiaddr.Multiaddr) ( relayID peer.ID ){
	smaddr := MaddrToStringList(mas)
	for _, saddr := range smaddr {
		_, rID, _ := MaddrRelayParse(saddr)

		if rID.String() != "" {
			relayID = rID
		}
	}

	return
}
