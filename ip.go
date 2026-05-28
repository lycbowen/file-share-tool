package main

import (
	"errors"
	"net"
	"strconv"
	"time"
)

func GetInternalIP() (string, error) {
	if ip, err := getOutboundIPv4(); err == nil {
		return ip.String(), nil
	}

	ip, err := getInterfaceIPv4()
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

func getOutboundIPv4() (net.IP, error) {
	dialer := net.Dialer{Timeout: 200 * time.Millisecond}
	conn, err := dialer.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || udpAddr.IP == nil {
		return nil, errors.New("local address is not UDP IPv4")
	}
	ip := udpAddr.IP.To4()
	if !isUsableIPv4(ip) {
		return nil, errors.New("default route has no usable IPv4 address")
	}
	return ip, nil
}

func getInterfaceIPv4() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var best net.IP
	bestPriority := 99
	for _, iface := range ifaces {
		if !isUsableInterface(iface) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		ip := bestIPv4FromAddrs(addrs)
		if ip == nil {
			continue
		}
		priority := ipv4Priority(ip)
		if best == nil || priority < bestPriority {
			best = ip
			bestPriority = priority
		}
	}

	if best != nil {
		return best, nil
	}
	return nil, errors.New("no active non-loopback IPv4 address found")
}

func isUsableInterface(iface net.Interface) bool {
	return iface.Flags&net.FlagUp != 0 &&
		iface.Flags&net.FlagLoopback == 0 &&
		iface.Flags&net.FlagPointToPoint == 0
}

func bestIPv4FromAddrs(addrs []net.Addr) net.IP {
	var best net.IP
	bestPriority := 99
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP == nil || ipNet.IP.IsLoopback() {
			continue
		}
		ip := ipNet.IP.To4()
		if !isUsableIPv4(ip) {
			continue
		}
		priority := ipv4Priority(ip)
		if best == nil || priority < bestPriority {
			best = ip
			bestPriority = priority
		}
	}
	return best
}

func isUsableIPv4(ip net.IP) bool {
	return ip != nil &&
		!ip.IsLoopback() &&
		!ip.IsUnspecified() &&
		!isBenchmarkIPv4(ip)
}

func ipv4Priority(ip net.IP) int {
	if isPrivateIPv4(ip) {
		return 0
	}
	if isLinkLocalIPv4(ip) {
		return 2
	}
	return 1
}

func isPrivateIPv4(ip net.IP) bool {
	return len(ip) == net.IPv4len && (ip[0] == 10 ||
		(ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31) ||
		(ip[0] == 192 && ip[1] == 168))
}

func isLinkLocalIPv4(ip net.IP) bool {
	return len(ip) == net.IPv4len && ip[0] == 169 && ip[1] == 254
}

func isBenchmarkIPv4(ip net.IP) bool {
	return len(ip) == net.IPv4len && ip[0] == 198 && (ip[1] == 18 || ip[1] == 19)
}

func buildServerURL(host string, port int) string {
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port))
}
