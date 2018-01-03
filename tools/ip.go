package tools

import (
	"fmt"
	"net"
)

// GetLocalIPS return list of local ips
func GetLocalIPS() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf(`Error while getting interface addrs: %v`, err)
	}

	ips := make([]net.IP, 0)

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}

	return ips, nil
}