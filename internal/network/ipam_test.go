package network

import (
	"net"
	"testing"
)

func Test_Allocate(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("192.168.0.0/24")
	ip, _ := IPAllocator.Allocate(ipnet)
	t.Logf("alloc ip: %v", ip)
}

func Test_Release(t *testing.T) {
	ip, ipnet, _ := net.ParseCIDR("192.168.0.1/24")
	IPAllocator.Release(ipnet, &ip)
}
