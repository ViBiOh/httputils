package tools

import (
	"net"

	"github.com/ViBiOh/httputils/pkg/errors"
)

// GetLocalIPS return list of local ips
func GetLocalIPS() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, errors.WithStack(err)
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

// GetLocalIP return first found local IP
func GetLocalIP() (net.IP, error) {
	ips, err := GetLocalIPS()
	if err != nil {
		return nil, err
	}

	if len(ips) > 0 {
		return ips[0], nil
	}

	return nil, nil
}
