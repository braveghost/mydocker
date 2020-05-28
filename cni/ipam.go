package cni

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"mydocker/setting"
	"net"
	"os"
	"strings"
)

func GetIpam() *Ipam {
	if defaultIpam == nil{
		defaultIpam = NewIpam()
	}
	return defaultIpam
}

func NewIpam()*Ipam{
	return &Ipam{
		map[string]*NetworkInfo{},
	}
}
var defaultIpam *Ipam



// 地址分配
// ip地址的管理
type Ipam struct {
	// 网段和ip位图, key 是网段或网络名, value是ip地址分配情况
	subnet map[string]*NetworkInfo
}

type NetworkInfo struct {
	Name string
	Ips  string
}

// 加载ip地址使用情况
func (i *Ipam) load() error {

	if _, err := os.Stat(setting.EContainerNetworkManagerDataPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(setting.EContainerNetworkManagerDataPath, 0622);err != nil{
				logrus.Errorf("Ipam.Load.NotExist | %v", err)
				return errors.WithMessage(err, "Ipam.Load.NotExist")
			}
		}else {
			logrus.Errorf("Ipam.Load.Stat | %v", err)
			return errors.WithMessage(err, "Ipam.Load.Stat")
		}
	}
	if _, err := os.Stat(setting.EContainerNetworkManagerDataFileName); err != nil {
		logrus.Warnf("Ipam.Load.Stat | %v | %s", err, setting.EContainerNetworkManagerDataFileName)
		return nil

	}
	f, err := os.Open(setting.EContainerNetworkManagerDataFileName)

	if err != nil {
		logrus.Errorf("Ipam.Load.OpenFile | %v", err)
		return errors.WithMessage(err, "Ipam.Load.OpenFile")
	}
	defer func() {
		_ = f.Close()
	}()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		logrus.Errorf("Ipam.Load.Read | %v | %s", err, string(content))
		return errors.WithMessage(err, "Ipam.Load.Read")

	}
	if err := json.Unmarshal(content, &i.subnet); err != nil {
		logrus.Errorf("Ipam.Load.Unmarshal | %v | %s", err, string(content))
		return errors.WithMessage(err, "Ipam.Load.Unmarshal")
	}

	return nil
}

// 存储ip地址信息
func (i *Ipam) dump() error {

	if _, err := os.Stat(setting.EContainerNetworkManagerDataPath); err != nil {
		if os.IsNotExist(err) {
			_ = os.MkdirAll(setting.EContainerNetworkManagerDataPath, 0644)
		} else {

			logrus.Errorf("Ipam.Dump.Stat | %v", err)
			return errors.WithMessage(err, "Ipam.Dump.Stat")
		}
	}

	f, err := os.OpenFile(setting.EContainerNetworkManagerDataFileName, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("Ipam.Dump.OpenFile | %v", err)

		return errors.WithMessage(err, "Ipam.Dump.OpenFile")
	}

	defer func() {
		_ = f.Close()
	}()
	data, _ := json.Marshal(i.subnet)
	if _, err := f.Write(data); err != nil {
		logrus.Errorf("Ipam.Dump.Marshal | %v", err)
		return errors.WithMessage(err, "Ipam.Dump.Marshal")
	}
	return nil
}

// 从网段中分配一个可用ip
func (i *Ipam) Allocator(subnet *net.IPNet) (net.IP, error) {
	//  加载网段信息
	if err := i.load(); err != nil {
		logrus.Errorf("Ipam.Dump.Marshal | %v", err)
		return nil, errors.WithMessage(err, "Ipam.Allocator.Load")
	}
	// 返回网段的子网掩码的总长度和网段前的固定位长度, 127.0.0.0/8, one为8, size为32
	one, size := subnet.Mask.Size()
	logrus.Infof("Allocator.Size | %d | %d", one, size)

	var (
		networkInfo *NetworkInfo
		ip          net.IP
		subip       = subnet.String()
	)
	// 判断网段是否存在，不存在就创建
	if _, exist := i.subnet[subip]; !exist {
		networkInfo = &NetworkInfo{
			Ips:  strings.Repeat("0", 1<<uint8(size-one)),
		}
		logrus.Infof("Allocator.Len | %s | %d", subip, len(networkInfo.Ips))
	} else {
		networkInfo = i.subnet[subip]
	}

	lenIps := len(networkInfo.Ips)
	for i, c := range networkInfo.Ips {
		// 匹配位图中的0位
		if c == '0' {
			ipalloc := []byte(networkInfo.Ips)
			ipalloc[i] = '1'
			networkInfo.Ips = string(ipalloc)
			ip = subnet.IP
			//
			//通过网段的 IP 与上面的偏移相加计算出分配的 IP 地址，由于 IP 地址是 uint 的一个数组，
			//要通过数组中的每一项加所需要的值，比如网段是 172. 16. 0. 0/12，数组序号是 65555.
			//那么在[172, 16, 0, 0]上依次加[uint8(65555 >> 24)、 uint8(65555 >> 16)、
			//uint8(65555 >> 8)、uint8(65555 >> 0)、即 [0, 1, 0, 19] 是 172.17.0.19
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(i >> ((t - 1) * 8))
			}
			if i == 0 || i == lenIps-1 {
				continue
			}
			break
		}
	}
	if ip == nil {
		return nil, errors.New("no ip can use")
	}
	i.subnet[subip] = networkInfo
	return ip, i.dump()

}

// 释放网络上的一个ip
func (i *Ipam) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	//  加载网段信息
	if err := i.load(); err != nil {
		logrus.Errorf("Ipam.Release.Load | %v", err)
		return errors.WithMessage(err, "Ipam.Release.Loadd")
	}

	var c int
	netStr := subnet.String()

	networkInfo, ok := i.subnet[netStr]
	logrus.Infof("Ipam.Release.Subnet | %v | %v",subnet, netStr)

	if !ok {
		logrus.Errorf("Ipam.Release.Subnet.NotOk | %v", netStr)
		return errors.New(fmt.Sprintf("Ipam network not exist: %s", netStr))
	}
	releaseIp := ipaddr.To4()

	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIp[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}
	logrus.Infof("Ipam.Dump.Marshal | %v | %s", c, netStr)
	mapip := []byte( networkInfo.Ips)
	mapip[c] = '0'
	networkInfo.Ips = string(mapip)
	return i.dump()

}
