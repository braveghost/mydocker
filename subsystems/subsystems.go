package subsystems

import (
	"bufio"
	"github.com/pkg/errors"
	"github.com/Sirupsen/logrus"
	"os"
	"path"
	"strings"
)

type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

type SubSystem interface {
	Name() string                               // 返回subsystem的名字, 如cpu
	Set(path string, res *ResourceConfig) error // 设置某个cgroup在这个subsystem中的资源限制
	Apply(path string, pid int) error           // 将进程添加到某个cgroup
	Remove(path string) error                   // 移除某个cgroup
}

var SubsystemsList = []SubSystem{
	&MemorySubSystem{},
	//&CpuSubSystem{},
}

// 得到cgroup在文件系统中的绝对路径
func GetCGroupPath(subsystem, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCGroupMountPoint(subsystem)
	p := path.Join(cgroupRoot, cgroupPath)
	_, err := os.Stat(p)
	if err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(p, 0755); err != nil {
				return "", errors.WithMessage(err, "error create cgroup")
			}
		}
		return p, nil
	}
	return "", errors.WithMessage(err, "cgroup path error")
}

func FindCGroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				logrus.Infof("%s | %v", opt, fields)
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logrus.Errorf("FindCGroupMountPoint.Scanner.Err | %s", err.Error())
	}
	return ""
}
