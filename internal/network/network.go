package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"godocker/internal/container"

	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

var (
	defaultNetworkPath = "/var/run/godocker/network/network/"
	drivers            = make(map[string]NetworkDriver)
	networks           = make(map[string]*Network)
)

type Network struct {
	Name    string     // network name
	IPRange *net.IPNet // ip address range
	Driver  string     // network driver name
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"port_mapping"`
	Network     *Network
}

func configEndpointIPAddressAndRoute(ep *Endpoint, containerInfo *container.Info) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	// 将容器的网络端点加入到容器的网络空间中.
	defer enterContainerNetns(&peerLink, containerInfo)()

	interfaceIP := *ep.Network.IPRange
	interfaceIP.IP = ep.IPAddress
	// logrus.Infof("interfaceIP: %v, interfaceIP.IP: %v", interfaceIP, interfaceIP.IP)

	if err := setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("set interface up %v,%s", ep.Network, err)
	}
	if err := setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	if err := setInterfaceUP("lo"); err != nil {
		return err
	}

	// 设置容器内的外部请求都通过容器内的 veth 端点访问
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	// 构建要添加的路由参数，包括网络设备、网关 IP 及目的网段
	// 相当于 route add -net 0.0.0.0/0 gw <bridge 网桥地址> dev <容器内的 Veth 端点设备>
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IPRange.IP,
		Dst:       cidr,
	}
	logrus.Infof("default Route: %v", defaultRoute)
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("netlink router add: %v", err)
	}

	return nil
}

func enterContainerNetns(erLink *netlink.Link, containerInfo *container.Info) func() {
	// 将容器的网络端点加入到容器的网络空间中
	// 并锁定当前程序所执行的线程，使当前线程进入到容器的网络空间
	// /proc/{pid}/ns/net
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", containerInfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("error get container net namespace, %v", err)
	}
	nsFD := f.Fd()

	// runtime.LockOSThread
	runtime.LockOSThread()

	if err := netlink.LinkSetNsFd(*erLink, int(nsFD)); err != nil {
		logrus.Errorf("error set link netns, %v", err)
	}

	origin, err := netns.Get()
	if err != nil {
		logrus.Errorf("error get current netns, %v", err)
	}

	if err := netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("error set netns, %v", err)
	}

	return func() {
		netns.Set(origin)
		origin.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configPortMapping(ep *Endpoint, containerInfo *container.Info) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("port mapping format error, %v", pm)
			continue
		}

		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables output, %v", output)
			continue
		}
	}

	return nil
}

func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}

	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("Error: %v", err)
		return err
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		logrus.Errorf("Network dump file name: %s, json marshal error: %v", nwPath, err)
		return err
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		logrus.Errorf("Network write content error: %v", err)
		return err
	}

	return nil
}

func (nw *Network) load(dumpPath string) error {
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}

	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		logrus.Errorf("Error load nw info %v", err)
		return err
	}

	return nil
}

// 检查文件状态，如果文件不存在就直接返回，反之进行删除。
func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return os.Remove(path.Join(dumpPath, nw.Name))
}

// CreateNetwork
// 1. 通过 IPAM 获取 IP.
// 2. 创建 network endpoint
// 3. 配置连接网络端点
// 4. 配置端口映射
func CreateNetwork(driver, subnet, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	gatewayIP, err := IPAllocator.Allocate(cidr) // todo: nw.IPRange
	if err != nil {
		return err
	}
	cidr.IP = net.IP(gatewayIP)
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(defaultNetworkPath)
}

func ConnectNetwork(networkName string, containerInfo *container.Info) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	// call ipam from network range get ip.
	ip, err := IPAllocator.Allocate(network.IPRange)
	if err != nil {
		return err
	}

	// create network endpoint
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", containerInfo.ID, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: containerInfo.PortMapping,
	}

	if err := drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}

	logrus.Infof("ep: %v, ip: %s", ep, ip)
	if err := configEndpointIPAddressAndRoute(ep, containerInfo); err != nil {
		return fmt.Errorf("config endpoint ip address and route error: %v", err)
	}

	return configPortMapping(ep, containerInfo)
}

func DisconnectNetwork(networkName string, containerInfo *container.Info) error {
	return nil
}

func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(w, "NAME\tIPRange\tDriver\n")

	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IPRange.String(),
			nw.Driver,
		)
	}

	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error: %v", err)
		return
	}
}

// DeleteNetwork
// 1. 删除网关IP
// 2. 删除网络对应的网络设备
// 3. 删除网络配置文件
func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}

	if err := IPAllocator.Release(nw.IPRange, &nw.IPRange.IP); err != nil {
		return fmt.Errorf("error remove network gateway ip: %v", err)
	}

	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("error remove network driver error: %s", err)
	}

	return nw.remove(defaultNetworkPath)
}

func Init() error {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}

	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &Network{
			Name: nwName,
		}

		// 加载失败后，应该退出本次循环.
		if err := nw.load(nwPath); err != nil {
			logrus.Errorf("Error load network: %s, error: %v", nw.Name, err)
		}

		networks[nwName] = nw
		return nil
	})

	return nil
}
