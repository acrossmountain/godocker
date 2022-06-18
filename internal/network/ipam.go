package network

import (
	"encoding/json"
	"net"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const ipamDefaultAllocatorPath = "/var/run/godocker/network/ipam/subnet.json"

var IPAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

type IPAM struct {
	SubnetAllocatorPath string
	Subnets             *map[string]string
}

// 加载网段地址分配信息
func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()

	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		logrus.Errorf("Error dump allocation info error: %v", err)
		return err
	}

	return nil
}

func (ipam *IPAM) dump() error {
	ipamCOnfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamCOnfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamCOnfigFileDir, 0644)
		} else {
			return err
		}
	}

	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()

	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}

	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	ipam.Subnets = &map[string]string{}

	err = ipam.load()
	if err != nil {
		logrus.Errorf("Error load allocation info, %v", err)
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	one, size := subnet.Mask.Size()

	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	for c := range (*ipam.Subnets)[subnet.String()] {
		// 找到数组中为 "0" 的项和数组序号，既可以分配的 IP
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			// 设置这个为 "0" 的序号值为 "1"，既分配这个 IP
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP

			// 通过网段的 IP 与上面的偏移相加计算出分配的 IP 地址，
			// 由于 IP 地址是 uint 的是一个数组，
			// 需要通过数组中的每每一项加所需要的值。
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}

			ip[3] += 1
			break
		}
	}

	ipam.dump()
	return
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipAddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		logrus.Errorf("Error dump allocation info error: %v", err)
	}

	// 计算 IP 地址在网段位图数组中的索引位置
	c := 0
	releaseIP := ipAddr.To4()
	releaseIP[3] -= 1

	for t := uint(4); t > 0; t -= 1 {
		// 与分配 IP 相反，释放 IP 获得索引的方式是 IP 地址的
		// 每一位相减之后分别左移将对应的数值加到索引上
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	ipAlloc := []byte((*ipam.Subnets)[subnet.String()])
	ipAlloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipAlloc)

	ipam.dump()
	return nil

}
