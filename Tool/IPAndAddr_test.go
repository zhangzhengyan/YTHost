package Tool

import (
	"fmt"
	"testing"
)

var localMa = "/ip4/0.0.0.0/tcp/9000"
var localMa2 = "/ip4/0.0.0.0/tcp/9001"
var localMa3 = "/ip4/0.0.0.0/tcp/9002"

func TestGetIps(t *testing.T) {
	fmt.Println(GetIPs())
}

func TestFilterIp4Addr(t *testing.T) {
	fmt.Println(FilterIp4Addr("/ip4"))
	fmt.Println(FilterIp4Addr("/ip6"))
}

func TestGetPortsFromMultiAddr(t *testing.T) {
	mas := []string{localMa, localMa2}
	fmt.Println(GetPortsFromMultiAddr(mas))
}

func TestGetMultiAddr(t *testing.T) {
	mas := []string{localMa, localMa2}
	fmt.Println(GetMultiAddr(mas))
}