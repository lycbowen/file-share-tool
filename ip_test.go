package main

import (
	"net"
	"testing"
)

func TestBestIPv4FromAddrsPrefersPrivateIPv4(t *testing.T) {
	addrs := []net.Addr{
		ipNet("169.254.1.2"),
		ipNet("203.0.113.10"),
		ipNet("192.168.1.8"),
	}

	got := bestIPv4FromAddrs(addrs)
	if got == nil || got.String() != "192.168.1.8" {
		t.Fatalf("bestIPv4FromAddrs() = %v, want 192.168.1.8", got)
	}
}

func TestBestIPv4FromAddrsIgnoresLoopbackUnspecifiedAndIPv6(t *testing.T) {
	addrs := []net.Addr{
		ipNet("127.0.0.1"),
		ipNet("0.0.0.0"),
		ipNet("fe80::1"),
		ipNet("10.0.0.5"),
	}

	got := bestIPv4FromAddrs(addrs)
	if got == nil || got.String() != "10.0.0.5" {
		t.Fatalf("bestIPv4FromAddrs() = %v, want 10.0.0.5", got)
	}
}

func TestBestIPv4FromAddrsIgnoresBenchmarkNetwork(t *testing.T) {
	addrs := []net.Addr{
		ipNet("198.18.0.1"),
		ipNet("198.19.255.254"),
		ipNet("192.168.5.87"),
	}

	got := bestIPv4FromAddrs(addrs)
	if got == nil || got.String() != "192.168.5.87" {
		t.Fatalf("bestIPv4FromAddrs() = %v, want 192.168.5.87", got)
	}
}

func TestIsUsableInterface(t *testing.T) {
	tests := []struct {
		name  string
		flags net.Flags
		want  bool
	}{
		{name: "active ethernet or wifi", flags: net.FlagUp | net.FlagBroadcast, want: true},
		{name: "down", flags: net.FlagBroadcast, want: false},
		{name: "loopback", flags: net.FlagUp | net.FlagLoopback, want: false},
		{name: "point to point", flags: net.FlagUp | net.FlagPointToPoint, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUsableInterface(net.Interface{Flags: tt.flags})
			if got != tt.want {
				t.Fatalf("isUsableInterface() = %t, want %t", got, tt.want)
			}
		})
	}
}

func ipNet(raw string) *net.IPNet {
	ip := net.ParseIP(raw)
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
}
