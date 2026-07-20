package server

import (
	"fmt"
	"net"
)

// LANAddrs returns http:// base URLs for each non-loopback IPv4 address
// found on the host's network interfaces, so a LAN client can paste one
// directly (e.g. into VLC) without looking up the machine's IP.
func LANAddrs(port int) []string {
	var addrs []string

	ifaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return addrs
	}

	for _, a := range ifaceAddrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok {
			continue
		}
		if ipNet.IP.To4() == nil {
			continue
		}
		if ipNet.IP.IsLoopback() {
			continue
		}
		addrs = append(addrs, fmt.Sprintf("http://%s:%d", ipNet.IP.String(), port))
	}

	return addrs
}
