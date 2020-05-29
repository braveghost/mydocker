package command

import (
	"mydocker/container"
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"mydocker/cgroup"
	"mydocker/cni"
	"mydocker/images"
	"mydocker/subsystems"
)

// 运行容器的参数
type RunArgs struct {
	Tty            bool
	CmdArray       []string
	Image          string
	Volume         string
	Name           string
	Env            []string
	NetworkName    string
	Ports          []string
	ResourceConfig *subsystems.ResourceConfig
}

/*
运行一个容器
@params args: 命令行参数
@return error:
*/
func Run(args *RunArgs) {
	logrus.Infof("Run.Params | %v | %v | %v | %s", args.Tty, args.CmdArray, args.ResourceConfig, args.Image)
	// 获取容器元信息
	meta, _ := container. GetContainerMeta(args.Name)

	if meta != nil {
		logrus.Warnf("Run.ContainerExist | %v | %v | %v | %s | %v", args.Tty, args.CmdArray, args.ResourceConfig, args.Image, meta)
		return
	}
	// 创建父进程对象
	parent, writePipe := container.NewParentProcess(args.Tty, args.Volume, args.Image, args.Name, args.Env)
	if parent == nil || writePipe == nil {
		logrus.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Errorf("Run.Error | %v | %v", err, parent)
		return
	}

	// 设置资源限制 todo 待测 没stress命令
	//cgroupManger := SetCGroup(args.Name, parent.Process.Pid, args.ResourceConfig)
	//defer func() {
	//	if err := cgroupManger.Destroy(); err != nil {
	//		logrus.Errorf("Run.Destroy | %v", err)
	//	}
	//}()

	// 生成 元数据对象
	meta = container.NewContainerMeta(parent.Process.Pid, args.Name, args.Image, args.Volume, args.CmdArray, args.Ports)
	if len(args.NetworkName) != 0 {
		// 链接网络
		_ = cni.InitNetworkList()
		meta.Network = args.NetworkName
		err := cni.ConnectNetwork(meta)
		if err != nil {
			logrus.Errorf("Run.ConnectNetwork | %v", err)
			return
		}
	}

	// 记录元数据
	if err := container.RecordContainerMeta(meta); err != nil {
		logrus.Errorf("Run.RecordContainerMeta | %v", err)
		return
	}

	// 发送初始化命令
	sendInitCommand(args.CmdArray, writePipe)
	if args.Tty {
		// ti参数前台等待
		// 前台等待台运行
		logrus.Infof("Run.Wait | %v", parent.Wait())
		_, workUrl := images.GetWriteWorkLayerOverlay(args.Name)
		// 删除挂载卷
		container.DeleteWorkSpace(workUrl, args.Volume)
		// 更新容器状态
		_ = container.UpdateContainerStatus(args.Name, container.STOP)
		//DeleteContainerMeta(cname)
	}
}

func SetCGroup(name string, pid int, res *subsystems.ResourceConfig) *cgroup.CGroupManager {
	cgroupManger := cgroup.NewCGroupManager(name)

	if err := cgroupManger.Set(res); err != nil {
		logrus.Errorf("Run.Set | %v", err)
	}
	logrus.Infof("Run.Set.Info")

	if err := cgroupManger.Apply(pid); err != nil {
		logrus.Errorf("Run.Apply | %v", err)

	}
	logrus.Infof("Run.Set.Apply")
	return cgroupManger
}

/*
发送命令字符串到容器进程
@params comArray: 命令
@params writePipe: 接收进程的文件句柄
*/
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	logrus.Infof("command all is %s", command)
	n, err := writePipe.WriteString(command)
	if err != nil {
		logrus.Errorf("sendInitCommand.ERROR | %v | %v", n, err)
	}
	logrus.Infof("sendInitCommand.Info | %v", writePipe.Close())
}
