package container

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"mydocker/images"
	"mydocker/setting"
	"mydocker/subsystems"
)

/*
读取命令管道命令
@return []string: 命令行参数
*/
func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		logrus.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	logrus.Info("readUserCommand | %s", msgStr)
	return strings.Split(msgStr, " ")
}

/*
通过初始化一个容器
@return error:
*/
func RunContainerInitProcess() error {
	// 读取命令管道命令
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return errors.New("Run container get user command error, cmdArray is nil")
	}

	// 读取命令管道命令
	subsystems.SetUpMount()

	// 查看命令是否存在
	p, err := exec.LookPath(cmdArray[0])
	if err != nil {
		logrus.Errorf("Exec loop path error %v", err)
		return err
	}
	logrus.Infof("Find path %s | %v", p, cmdArray)
	// 执行命令
	if err := syscall.Exec(p, cmdArray[0:], os.Environ()); err != nil {
		logrus.Errorf("RunContainerInitProcess.Exec.Error | %s", err.Error())
	}

	return nil
}

/*
运行一个容器
@params args: 命令行参数
@return error:
*/
func NewParentProcess(tty bool, volume, image, name string, envList []string) (*exec.Cmd, *os.File) {
	// 新建命令管道
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		logrus.Errorf("New pipe error %v", err)
		return nil, nil
	}
	// 自执行初始化容器
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 加载环境变量，宿主机环境+参数环境
	cmd.Env = append(os.Environ(), envList...)
	if tty {
		// 前台运行
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

	} else {
		//	后台运行时候日志存储到文件
		logDir := GetLogPath(name)
		if err := os.MkdirAll(logDir, 0622); err != nil {
			logrus.Errorf("NewParentProcess.MkdirAll.Log | %v | %s", err, logDir)
			return nil, nil
		}
		logFile := path.Join(logDir, setting.EContainerLogName)
		slf, err := os.Create(logFile)
		if err != nil {
			logrus.Errorf("NewParentProcess.Create.Log | %v | %s", err, logFile)
			return nil, nil
		}
		cmd.Stdout = slf
	}
	// 扩展文件描述符
	cmd.ExtraFiles = []*os.File{readPipe}
	// images.NewWorkSpaceAufs
	// aufs是老版本了，所以就没试过了，这里只用了overlay2
	upperDir, workDir := images.NewWorkSpaceOverlay(image, name)
	logrus.Infof("NewParentProcess.CreateMountPointOverlay | %s | %s", upperDir, workDir)
	// 设置工作路径
	cmd.Dir = workDir
	// 挂载卷信息
	NewWorkSpace(volume, name)
	return cmd, writePipe
}

/*
拼接日志路径
@params name: 容器名称
@return string: 日志文件路径
*/
func GetLogPath(name string) string {
	return path.Join(setting.EContainerLogsDataPath, name)

}

/*
新建命令管道
@params args: 命令行参数
@return *os.File: 读取端
@return *os.File: 写入端
@return error:
*/
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

/*
删除挂载卷
@params mntUrl: 容器数据路径
@return volume: 卷路径
*/
func DeleteWorkSpace(mntUrl, volume string) {
	if len(volume) != 0 {
		// 解析卷路径
		vu := volumesUrlExtract(volume)
		if len(vu) != 2 {
			logrus.Errorf("NewWorkSpace.Split.Len | %v", volume)
			return
		}
		left, right := vu[0], vu[1]
		if len(left) != 0 && len(right) != 0 {
			cvu := path.Join(mntUrl, right)
			logrus.Infof("umount %s", cvu)
			// 卸载
			cmd := exec.Command("umount", cvu)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				logrus.Errorf("DeleteWorkSpace.Command.Run | %v", err)
			}
		}
	}

}

/*
挂载卷
@params volume: 卷路径
@return name: 容器名
*/
func NewWorkSpace(volume, name string) {
	_, workDir := images.GetWriteWorkLayerOverlay(name)
	if len(volume) != 0 {
		// 解析卷信息
		vu := volumesUrlExtract(volume)
		if len(vu) != 2 {
			logrus.Errorf("NewWorkSpace.Split.Len | %v", volume)
			return
		}
		left, right := vu[0], vu[1]
		if len(left) != 0 && len(right) != 0 {
			// 挂载卷
			MountVolume(workDir, left, right)
		}
	}
}

/*
解析卷路径
@params volume: 卷路径
*/
func volumesUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}

/*
挂载卷逻辑
@params mntUrl: 容器数据层路径
@params left: 宿主机路径
@params right: 容器路径
*/
func MountVolume(mntUrl, left, right string) {
	if err := os.Mkdir(left, 0777); err != nil {
		logrus.Errorf("MountVolume.Mkdir.Left | %s | %v", left, err)
	}
	// 容器挂载点路径
	cvu := path.Join(mntUrl, right)
	if err := os.Mkdir(cvu, 0777); err != nil {
		logrus.Errorf("MountVolume.Mkdir.Right | %s | %v", right, err)
	}
	// 把宿主机文件目录挂载到容器挂载点
	logrus.Infof("mount %s %s", cvu, left)
	cmd := exec.Command("mount", "--bind", left, cvu)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("MountVolume.Command.Run | %v", err)
	}
}
