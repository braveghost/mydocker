package cni

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

var (
	ErrCreateNetInterface = errors.New("create net interface error")
	ErrNetInterfaceName   = errors.New("bridge interface name error")
)

type BridgeNetworkDriver struct{}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (d *BridgeNetworkDriver) Create(subnet, name string) (*Network, error) {

	// 获取到网关地址和网络ip段
	ip, cidr, _ := net.ParseCIDR(subnet)
	cidr.IP = ip
	// 初始化网络对象
	return d.initBridge(&Network{
		Name:    name,
		IpRange: cidr,
		Driver:  d.Name(),
	})
}

func (d *BridgeNetworkDriver) initBridge(n *Network) (*Network, error) {
	name := n.Name
	// 创建bridge虚拟设备
	if err := createBridgeInterface(name); err != nil {
		logrus.Errorf("BridgeNetworkDriver.InitBridge.CreateBridgeInterface | %v", err)
		return nil, err
	}
	// 设置bridge设备地址和路由, 并启动设备
	gatewayIp := *n.IpRange
	gatewayIp.IP = n.IpRange.IP
	if err := setInterfaceIp(name, gatewayIp.String(), true); err != nil {
		logrus.Errorf("BridgeNetworkDriver.InitBridge.SetInterfaceIp | %v", err)
		return nil, err
	}
	// 设置iptables snat规则
	if err := setIptablesRole(name, n.IpRange); err != nil {
		logrus.Errorf("BridgeNetworkDriver.InitBridge.SetIptablesRole | %v", err)
		return nil, err
	}
	return n, nil
}

func (d *BridgeNetworkDriver) Delete(network *Network) error {
	//	 网络名即设备名
	//	 查找设备
	br, err := netlink.LinkByName(network.Name)
	if err != nil {
		return errors.WithMessage(err, "BridgeNetworkDriver.Delete.LinkByName")
	}
	//	 删除对应的设备
	if err := netlink.LinkDel(br); err != nil {
		return errors.WithMessage(err, "BridgeNetworkDriver.Delete.LinkDel")
	}
	return nil
}

func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	br, err := netlink.LinkByName(network.Name)
	if err != nil {
		return errors.WithMessage(err, "BridgeNetworkDriver.Connect.LinkByName")
	}

	// 创建 veth 接口对象
	la := netlink.NewLinkAttrs()
	// 长度限制
	la.Name = endpoint.ID[:10]

	// 设置 veth 的 master 属性，设置这个 veth 的一段到网络的 bridge 上 (这里是挂载网桥)
	la.MasterIndex = br.Attrs().Index

	// 创建veth，通过PeerName配置另外一端的端口名
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + la.Name,
	}

	// 真正的创建 veth 接口
	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return errors.WithMessage(err, "BridgeNetworkDriver.Connect.LinkAdd")
	}
	// 设置veth启动
	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		return errors.WithMessage(err, "BridgeNetworkDriver.Connect.LinkSetUp")
	}
	return nil
}

func (d *BridgeNetworkDriver) DisConnect(network *Network, endpoint *Endpoint) error {
	panic("implement me")
}

/*
创建bridge设备
@params name: 网络设备名称, 最长15位
@return error:
*/
func createBridgeInterface(name string) error {
	if len(name) == 0 || len(name) > 15 {
		return ErrNetInterfaceName
	}
	// 检查是否存在同名设备
	_, err := net.InterfaceByName(name)
	if err == nil {
		// 已经存在
		return errors.WithMessage(ErrCreateNetInterface, "already existed")
	}
	// 除了不存在外的其他的错误
	if !strings.Contains(err.Error(), "no such network interface") {
		return errors.WithMessage(ErrCreateNetInterface, err.Error())
	}

	//	todo 不存在则新建

	// 创建link对象
	la := netlink.NewLinkAttrs()
	la.Name = name
	// 创建bridge对象
	n := &netlink.Bridge{LinkAttrs: la}

	// 创建设备
	if err := netlink.LinkAdd(n); err != nil {
		logrus.Errorf("CreateBridgeInterface.LinkAdd | %v | %v", err, n)
		return errors.WithMessage(err, "CreateBridgeInterface.LinkAdd")
	}
	return nil
}

/*
设置网络接口ip及路由转发
@params name: 网络设备名称
@params rawIp: 要设置的ip地址
@params up: true表示配置完后启动网络
@return error:
*/
func setInterfaceIp(name, rawIp string, up bool) error {
	// 通过设备名获取网络接口
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return errors.WithMessage(err, "SetInterfaceIp.LinkByName")
	}

	// 解析cidr
	ipNet, err := netlink.ParseIPNet(rawIp)
	if err != nil {
		return errors.WithMessage(err, "SetInterfaceIp.ParseIPNet")
	}

	// 配置地址，并且配置路由表将网段转发到指定网络接口上
	if err := netlink.AddrAdd(iface, &netlink.Addr{IPNet: ipNet}); err != nil {
		return errors.WithMessage(err, "SetInterfaceIp.AddrAdd")
	}

	if up {
		return startInterface(iface)
	}

	return nil
}

/*
通过名称启动网络接口
@params name: 网络设备名称
@return error:
*/
func startInterfaceByName(name string) error {
	// 通过设备名获取网络接口
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return errors.WithMessage(err, "StartInterfaceByName.LinkByName")
	}
	// 启动网络接口
	return startInterface(iface)
}

/*
启动网络接口
@params link: 网络接口对象
@return error:
*/
func startInterface(link netlink.Link) error {
	if err := netlink.LinkSetUp(link); err != nil {
		return errors.WithMessage(err, "StartInterface.LinkSetUp")
	}
	return nil
}

/*
设置iptables转发规则，对应的bridge的MASQUERADE规则
@params name: 网络设备名称
@params subnet: 网络接口对象
@return error:
*/
func setIptablesRole(name string, subnet *net.IPNet) error {

	// todo 可以使用库 github.com/coreos/go-iptables

	// iptables -t nat -A POSTROUTING -s <bridgeNarne> ! -o <bridgeNarne> -] MASQUERADE

	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), name)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		logrus.Errorf("SetIptablesRole.Output | %v | %s", err, string(output))
	}
	return nil
}
