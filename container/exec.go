package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	_ "mydocker/nsenter"
	"os"
	"os/exec"
	"strings"
)

const (
	ENV_EXEC_PID = "mydocker_pid"
	ENV_EXEC_CMD = "mydocker_cmd"
)

func ExecContainer(name string, command []string) error {
	meta, err := GetContainerMeta(name)
	if err != nil {
		return err
	}
	cmdStr := strings.Join(command, " ")
	logrus.Infof("ExecContainer.Info | %s | %v", cmdStr, meta)

	// 重新执行当前命令，因为一开始执行时候并没有触发到c的切换ns代码，设置环境变量后重新fork执行触发
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	_ = os.Setenv(ENV_EXEC_CMD, cmdStr)
	_ = os.Setenv(ENV_EXEC_PID, meta.Pid)
	//defer func() {
	//	_ = os.Setenv(ENV_EXEC_CMD, "")
	//	_ = os.Setenv(ENV_EXEC_PID, "")
	//
	//}()
	cmd.Env = append(os.Environ(), GetEnvByPid(meta.Pid)...)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("ExecContainer.Run | %v", err)
	}

	return nil
}

// 通过pid获取进程内环境变量
func GetEnvByPid(pid string) []string {
	// 进程的环境变量存放在/proc/PID/environ文件
	content, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/environ", pid))
	if err != nil {
		logrus.Errorf("GetEnvByPid.ReadFile | %v", err)
		return nil
	}
	// 环境变量分隔符是\0000
	return strings.Split(string(content), "\u0000")
}
