package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

// Name return driver name
func (bnd *BridgeNetworkDriver) Name() string {
	return "bridge"
}

// Create create network
func (bnd *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip

	n := &Network{
		Name:    name,
		IPRange: ipRange,
		Driver:  bnd.Name(),
	}

	// 配置 Linux Bridge
	err := bnd.initBridge(n)
	if err != nil {
		logrus.Errorf("Error init bridge: %v", err)
	}

	return n, err
}

// 1. 创建 bridge 虚拟设备
// 2. 设置 bridge 设备地址和路由
// 3. 启动 bridge 设备
// 4. 设置 iptables SNAT 规则
func (bnd *BridgeNetworkDriver) initBridge(network *Network) error {
	bridgeName := network.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("error create bridge: %s, error: %v", bridgeName, err)
	}

	gatewayIP := *network.IPRange
	gatewayIP.IP = network.IPRange.IP
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return fmt.Errorf("error assign ip address: %s, on bridge: %s with an error of: %v", &gatewayIP, bridgeName, err)
	}

	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("error set bridge up: %s,Error: %v", bridgeName, err)
	}

	if err := setupIPTables(bridgeName, network.IPRange); err != nil {
		return fmt.Errorf("error setting iptables for %s: %v", bridgeName, err)
	}

	return nil
}

// Delete delete network
func (bnd *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	return netlink.LinkDel(iface)
}

// Connect connect container network to endpoint network
func (bnd *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	iface, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = iface.Attrs().Index

	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	// 上面指定了 link 的 MasterIndex 是网路对应的 linux bridge.
	// veth 的一端就已经挂载到了网络对应的 linux bridge 上.
	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error add endpoint device: %v", err)
	}

	// 调用 netlink 的 LinkSetUp 方法，设置 veth 启动
	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error set endpoint device up: %v", err)
	}

	return nil
}

// Disconnect
func (bnd *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	return nil
}

// 创建 linux bridge 设备
func createBridgeInterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	// 初始化一个 netlink 的 Liunx 基础对象
	// Link 的名字即 Bridge 虚拟设置的名字
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	// 调用 netlink 的 Linkadd 方法，创建 Bridge 虚拟网络设备
	br := &netlink.Bridge{
		LinkAttrs: la,
	}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge create failed for bridge %s: %v", bridgeName, err)
	}

	return nil
}

// 设置 Bridge 设备的网络地址和路由
func setInterfaceIP(name, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		logrus.Debugf("error retrieving new bridge netlink link [ %s ]... retrying", name)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot the error: %v", err)
	}
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0, Broadcast: nil}
	return netlink.AddrAdd(iface, addr)
}

// 设置网络接口为 UP 状态
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("error retrieving a link named [ %s ]: %v",
			iface.Attrs().Name, err)
	}

	// ip link set xx up
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("error enable interface for %s: %v", interfaceName, err)
	}

	return nil
}

// 设置 iptables 对应 bridge 的 MASQUERADE 规则
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// iptables -t nat -A POSTROUTING -s <bridgeName> ! -o <bridgeName> -j MASQUERADE
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE",
		subnet.String(), bridgeName,
	)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("iptables output, %v", output)
	}
	return nil
}
