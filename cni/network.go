package cni

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"mydocker/container"
	"mydocker/setting"
	"mydocker/utils"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

var (
	ErrNotFoundNetwork = errors.New("not found network")
	ErrPortMapping     = errors.New("port mapping error")
)

/*
创建网络
@params driver: 网络驱动
@params subnet: cidr ip , 如 127.0.0.1/24
@params name: 网络名称
@return error:
*/
func CreateNetwork(driver, subnet, name string) error {
	// 将字符串ip转换成ip对象
	ip, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		logrus.Errorf("CreateNetwork.ParseCID | %v | %s", err, ip)
		return errors.WithMessage(err, "CreateNetwork.ParseCIDR")
	}
	logrus.Infof("CreateNetwork.ParseCID.Info | %v | %v", ip, cidr)

	// 获取一个可用ip，作为网关
	gatewayip, err := GetIpam().Allocator(cidr)
	if err != nil {
		logrus.Errorf("CreateNetwork.Allocator | %v | %v", err, gatewayip)
		return errors.WithMessage(err, "CreateNetwork.Allocator")
	}
	cidr.IP = gatewayip

	// 调用驱动创建网络
	nw, err := GetNetworkDriverManager().Get(driver).Create(cidr.String(), name)
	if err != nil {
		logrus.Errorf("CreateNetwork.Create | %v | %s", err, cidr.String())

		return errors.WithMessage(err, "CreateNetwork.Create")
	}
	// 保存网络信息
	return nw.dump()
}

// 链接网络
func ConnectNetwork(name string, meta *container.ContainerMeta) error {
	nw, ok := networks[name]
	if !ok {
		logrus.Errorf("ConnectNetwork.Get.Network | %v", ErrNotFoundNetwork)

		return errors.WithMessage(ErrNotFoundNetwork, "ConnectNetwork")
	}

	_, cidr, err := net.ParseCIDR(nw.IpRange.String())
	if err != nil {
		logrus.Errorf("ConnectNetwork.ParseCID | %v", err)
		return errors.WithMessage(err, "DeleteNetwork.ParseCIDR")
	}
	logrus.Infof("ConnectNetwork.ParseCID.Info | %v", cidr)


	// 获取一个ip地址
	ip, err := GetIpam().Allocator(cidr)
	if err != nil {
		logrus.Errorf("ConnectNetwork.Allocator | %v", err)

		return errors.WithMessage(err, "ConnectNetwork.Allocator")
	}
	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", meta.Id, name),
		IpAddress:   ip,
		MacAddress:  net.HardwareAddr("MAC-" + utils.GetSnowId()),
		PortMapping: meta.PortMapping,
		Network:     nw,
	}
	logrus.Infof("ConnectNetwork.Info | %v | %v", ip, nw)
	// 调用网络驱动链接网络
	if err := defaultNetworkDriverManager.Get(nw.Driver).Connect(nw, ep); err != nil {
		logrus.Errorf("ConnectNetwork.Connect | %v", err)
		return errors.WithMessage(err, "ConnectNetwork.Connect")
	}

	// 进入到容器的namespace 配置ip和路由
	if err := configEndpointIpAddressAndRouter(ep, meta); err != nil {
		logrus.Errorf("ConnectNetwork.ConfigEndpointIpAddressAndRouter | %v", err)
		return errors.WithMessage(err, "ConnectNetwork.ConfigEndpointIpAddressAndRouter")
	}
	meta.Ip = ip.To4().String()
	// 配置端口映射
	configPortMapping(ep, meta)
	return nil
}

/*
配置容器的网络端点地址和路由
@params ep: 容器网络端点信息
@params meta: 容器的元信息
@return error:
*/
func configEndpointIpAddressAndRouter(ep *Endpoint, meta *container.ContainerMeta) error {
	// 拿到容器使用的 veth 接口端
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		logrus.Errorf("ConfigEndpointIpAddressAndRouter.LinkByName | %v", err)
		return errors.WithMessage(err, "ConfigEndpointIpAddressAndRouter.LinkByName")
	}
	// 将接口加入容器网络空间，并且换进程到容器的网络空间
	if fn, err := enterContainerNetns(&peerLink, meta); err != nil {
		return errors.WithMessage(err, "ConfigEndpointIpAddressAndRouter.EnterContainerNetns")
	} else {
		//defer执行函数时候退出到宿主机的网络空间,即下面的代码都是在容器网络空间执行的
		defer fn()
	}

	// 设置容器ip地址并启动
	interfaceIp := *ep.Network.IpRange
	interfaceIp.IP = ep.IpAddress
	logrus.Infof("ConfigEndpointIpAddressAndRouter.Info | %v | %v | %v |%v |%v",interfaceIp, ep.Device.PeerName, ep.IpAddress.String(),interfaceIp.String())
	if err = setInterfaceIp(ep.Device.PeerName, interfaceIp.String(), true); err != nil {
		logrus.Errorf("ConfigEndpointIpAddressAndRouter.SetInterfaceIp | %v | %v", err, interfaceIp.String())

		return errors.WithMessage(err, "ConfigEndpointIpAddressAndRouter.SetInterfaceIp")
	}
	// 启动lo网卡
	if err = startInterfaceByName("lo"); err != nil {
		logrus.Errorf("ConfigEndpointIpAddressAndRouter.StartInterfaceByName | %v", err)
		return errors.WithMessage(err, "ConfigEndpointIpAddressAndRouter.StartInterfaceByName")

	}
	// 设置容器内的外部请求都通过容器内的veth端点访问
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	// 构建路由
	// ip route add default dev veth0
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		logrus.Errorf("ConfigEndpointIpAddressAndRouter.RouteAdd | %v", err)
		return errors.WithMessage(err, "ConfigEndpointIpAddressAndRouter.RouteAdd")
	}

	return nil
}

/*
将接口加入容器网络空间，并且换进程到容器的网络空间，返回的函数为退出到宿主机的网络空间
@params enLink: 容器内的网络接口对象
@params meta: 容器的元信息
@return func(): 函数为退出到宿主机的网络空间
*/
func enterContainerNetns(enLink *netlink.Link, meta *container.ContainerMeta) (func(), error) {
	//	通过容器的pid， 找到容器的net namespace
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", meta.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("EnterContainerNetns.OpenFile | %v", err)
		return nil, errors.WithMessage(err, "EnterContainerNetns.OpenFile")
	}

	//	获取文件描述符
	nsFD := f.Fd()
	// 锁定当前程序执行的线程, 不锁定的话goroutine可能会被调度到别的线程，不能保证一直在所需要的网络空间中
	runtime.LockOSThread()
	// 修改veth的另一端，将其移动到容器的net namespace
	if err := netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("EnterContainerNetns.LinkSetNsFd | %v", err)
	}

	//	获取当前网络（宿主机）的net namespace，为了回退
	origns, err := netns.Get()
	if err != nil {
		logrus.Errorf("EnterContainerNetns.GetNs | %v", err)

	}

	// 切换当前进程到对应的net namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("EnterContainerNetns.SetNs | %v", err)

	}
	return func() {
		// 回退到原来的net namespace
		_ = netns.Set(origns)
		// 关闭namespace
		_ = origns.Close()
		//解锁线程锁定
		runtime.UnlockOSThread()
		// 关闭namespace文件
		_ = f.Close()
	}, nil

}

/*
绑定端口映射
@params ep: 容器网络端点信息
@params meta: 容器的元信息
*/
func configPortMapping(ep *Endpoint, meta *container.ContainerMeta) {
	for _, pm := range ep.PortMapping {
		//	 分割宿主机端口和容器端口
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("ConfigPortMapping.Split | %v", errors.WithMessage(ErrPortMapping, pm))
			continue

		}

		//	 配置映射关系
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IpAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("ConfigPortMapping.Output | %v | %s", err, string(output))
			continue
		}
	}
}

/*
遍历网络列表
*/
func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "NAME\tIP_RANGE\tDRIVER\n")

	for _, nw := range networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", nw.Name, nw.IpRange, nw.Driver)
	}

	if err := w.Flush(); err != nil {
		logrus.Errorf("ListNetwork | %v", err)
	}
}

/*
删除网络
@params name: 网络名称
@return error:
*/
func DeleteNetwork(name string) error {
	nt, ok := networks[name]
	if !ok {
		return errors.WithMessage(ErrNotFoundNetwork, "DeleteNetwork")
	}

	ip, cidr, err := net.ParseCIDR(nt.IpRange.String())
	if err != nil {
		logrus.Errorf("DeleteNetwork.ParseCID | %v | %s", err, ip)
		return errors.WithMessage(err, "DeleteNetwork.ParseCIDR")
	}
	logrus.Infof("DeleteNetwork.ParseCID.Info | %v | %v", ip, cidr)


	//	调用网络管理器释放网络网关ip
	if err := GetIpam().Release(cidr, &ip); err != nil {
		logrus.Errorf("DeleteNetwork.Release | %v | %v | %v", err, nt.IpRange,nt.IpRange.IP)
		return errors.WithMessage(err, "DeleteNetwork.Release")
	}

	//	删除网络设备
	if err := defaultNetworkDriverManager.Get(nt.Driver).Delete(nt); err != nil {
		logrus.Errorf("DeleteNetwork.Delete | %v", err)
		return errors.WithMessage(err, "DeleteNetwork.Delete")

	}
	// 删除配置文件
	if err := nt.remove(); err != nil {
		logrus.Errorf("DeleteNetwork.Remove | %v", err)
		return errors.WithMessage(err, "DeleteNetwork.Remove")
	}
	return nil
}


var networks = make(map[string]*Network)

/*
初始化网络 列表
@return error:
*/
func InitNetworkList() error {
	//	判断网络目录是否存在
	if _, err := os.Stat(setting.EContainerNetworkDataPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(setting.EContainerNetworkDataPath, 0644)

		}
		if err != nil {
			logrus.Errorf("InitNetworkList.Stat | %v", err)
			return errors.WithMessage(err, "InitNetworkList.Stat")
		}
	}

	//	 遍历目录中的所有文件

	return filepath.Walk(setting.EContainerNetworkDataPath, func(fp string, info os.FileInfo, err error) error {
		// 跳过目录
		if info.IsDir() {
			return nil
		}

		//	通过文件名加载网络
		_, name := path.Split(fp)
		nw := &Network{Name: name}

		if err := nw.load(); err != nil {
			logrus.Errorf("InitNetworkList.Walk.Load | %v", err)
			return errors.WithMessage(err, "InitNetworkList.Walk.Load")
		}
		networks[name] = nw
		return nil
	})

}
