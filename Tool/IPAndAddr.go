package Tool

import (
	addrutil "github.com/libp2p/go-addr-util"
	"strings"
)

// 从指定的multiAddr string 地址中获取端口
func GetPortsFromMultiAddr(mas []string) []string {
	ports := make([]string, 0)
	for _, ma := range mas {
		strs := strings.Split(ma, "/")
		if len(strs) < 5 || FilterIp4Addr(strs[1]) {
			continue
		} else {
			ports = append(ports, strs[4])
		}
	}
	return ports
}

// 选择ip4 的ip
func FilterIp4Addr(multiAddrIp string) bool {
	strs := strings.Split(multiAddrIp, "/")
	if len(strs) < 2 || !strings.EqualFold("ip4", strs[1]) {
		return false
	}
	return true
}

func GetIPs() []string {
	IPs := make([]string, 0)
	// 获取本机的ip地址(内网 + 外网地址)
	multiAddr, _ := addrutil.InterfaceAddresses()
	for _, ma := range multiAddr {
		// 转string
		maString := ma.String()
		// 过滤, 目前只选择ip4 的地址
		if !FilterIp4Addr(maString) {
			continue
		}
		IPs = append(IPs,maString)
	}
	return IPs
}

func GetMultiAddr(multiAddrs []string) []string {
	ports := GetPortsFromMultiAddr(multiAddrs)
	IPs := GetIPs()
	res := make([]string, 0)
	// 拼接
	for _, ip := range IPs {
		for _, port := range ports {
			ma := ip + "/tcp/" + port
			res = append(res, ma)
		}
	}
	return res
}