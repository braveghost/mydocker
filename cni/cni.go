package cni

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"io/ioutil"
	"mydocker/setting"
	"net"
	"os"
	"path"
)

// 容器网络
// 容器网络端点的载体
type Network struct {
	Name    string     // 网络名
	IpRange *net.IPNet // 网络地址段
	Driver  string     // 网络驱动
}

// 网络信息保存到文件
func (nw *Network) dump() error {
	// 检查网络文件目录
	if _, err := os.Stat(setting.EContainerNetworkDataPath); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(setting.EContainerNetworkDataPath, 0622); err != nil {
				logrus.Errorf("Network.Dump.MkdirAll | %v", err)
				return errors.WithMessage(err, "Network.Dump.MkdirAll")
			}
		} else {
			logrus.Errorf("Network.Dump.IsNotExist | %v", err)

			return errors.WithMessage(err, "Network.Dump.IsNotExist")
		}

	}
	//	文件名是网络名字
	nfp := path.Join(setting.EContainerNetworkDataPath, nw.Name)

	// 写数据到网络
	f, err := os.OpenFile(nfp, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("Network.Dump.OpenFile | %v", err)
		return errors.WithMessage(err, "Network.Dump.OpenFile")
	}
	defer func() {
		_ = f.Close()
	}()
	content, err := json.Marshal(nw)
	if err != nil {
		logrus.Errorf("Network.Dump.Marshal | %v", err)
		return errors.WithMessage(err, "Network.Dump.Marshal")
	}


	if _, err := f.Write(content); err != nil {
		logrus.Errorf("Network.Dump.Write | %v", err)

		return errors.WithMessage(err, "Network.Dump.Write")
	}
	logrus.Infof("Network.Dump.Write | %v", string(content))
	return nil
}

// 网络信息从文件加载
func (nw *Network) load() error {
	data, err := ioutil.ReadFile(path.Join(setting.EContainerNetworkDataPath, nw.Name))
	if err != nil {
		logrus.Errorf("Network.Load.ReadFile | %v", err)
		return errors.WithMessage(err, "Network.Load.ReadFile")
	}
	if err := json.Unmarshal(data, nw); err != nil {
		logrus.Errorf("Network.Load.Unmarshal | %v", err)
		return errors.WithMessage(err, "Network.Load.Unmarshal")
	}
	return nil
}


// 删除对应的网络配置文件
func (nw *Network) remove() error {
	f := path.Join(setting.EContainerNetworkDataPath, nw.Name)
	if _, err := os.Stat(f);err != nil{
		if os.IsNotExist(err){
			return nil
		}
		return errors.WithMessage(err, "Network.Remove.Stat")
	}

	return os.Remove(f)
}





// 容器网络端点
// 连接容器和网络的通信端点
type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"device"`
	IpAddress   net.IP           `json:"ip_address"`
	MacAddress  net.HardwareAddr `json:"mac_address"`
	PortMapping []string         `json:"port_mapping"`
	Network     *Network
}

// 网络驱动
// 网络的管理功能，创建连接销毁等
type INetworkDriver interface {
	Name() string                                          // 驱动名
	Create(subnet, name string) (*Network, error)          // 创建网络
	Delete(network *Network) error                         // 删除网络
	Connect(network *Network, endpoint *Endpoint) error    // 连接容器网络端点到网络
	DisConnect(network *Network, endpoint *Endpoint) error // 从网络上移除容器网络端点
}

