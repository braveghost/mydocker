package subsystems

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubSystem struct {
}

func (m MemorySubSystem) Name() string {
	return "memory"
}

func (m MemorySubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	subSysCgroupPath, err := GetCGroupPath(m.Name(), cgroupPath, true)
	if err == nil {
		if len(res.MemoryLimit) != 0 {
			//设置这个 cgroup 的内存限制，即将限制写入到 cgroup 对应目录的 memory.limit in bytes 文件中
			fp := path.Join(subSysCgroupPath, "memory.limit_in_bytes")
			if err := ioutil.WriteFile(fp, []byte(res.MemoryLimit), 0644); err != nil {
				return errors.WithMessage(err, "set cgroup memory fail")
			}
		}
	}
	return err
}

// 将一个进程加入到cgroupPath对应的cgroup中
func (m MemorySubSystem) Apply(cgroupPath string, pid int) error {
	subSysCgroupPath, err := GetCGroupPath(m.Name(), cgroupPath, true)
	if err != nil {
		return errors.WithMessage(err, "MemorySubSystem.Apply.GetCGroupPath.Error")
	}
	if err := ioutil.WriteFile(path.Join(subSysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return errors.WithMessage(err, "MemorySubSystem.Apply.WriteFile.Error")
	}
	return nil
}

// 删除cgroupPath对应的cgroup
func (m MemorySubSystem) Remove(cgroupPath string) error {
	subSysCgroupPath, err := GetCGroupPath(m.Name(), cgroupPath, false)
	if err != nil {
		return errors.WithMessage(err, "MemorySubSystem.Remove.GetCGroupPath.Error")
	}
	return os.Remove(subSysCgroupPath)
}

type CpuSubSystem struct {
}

func (c CpuSubSystem) Name() string {
	panic("implement me")
}

func (c CpuSubSystem) Set(path string, res *ResourceConfig) error {
	panic("implement me")
}

func (c CpuSubSystem) Apply(path string, pid int) error {
	panic("implement me")
}

func (c CpuSubSystem) Remove(path string) error {
	panic("implement me")
}
