package subsystems

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/pkg/errors"
	"os"
	"path/filepath"

	"syscall"
)


/*
挂载root
*/
func SetUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("SetUpMount.Getwd.Error | %v", err)
		return
	}
	logrus.Infof("Cuttent.Location | %s", pwd)

	if err := pivotRoot(pwd); err != nil {
		logrus.Errorf("SetUpMount.PivotRoot.Error | %v", err)

	}
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		logrus.Errorf("SetUpMount.Mount.Proc.Error | %v", err)
	}

	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		logrus.Errorf("SetUpMount.Mount.Tmpfs.Error | %v", err)

	}
}

func pivotRoot(root string) error {
	//pivot_root系统调用那里报错Invalid argument，然后程序退出后，/proc有问题,其他issue有提到，翻了下runc的代码，
	//发现是因为/这个mount point的标记位是share, 所以pivot_root切换rootfs失败，加上后面重新mount /proc的时候，
	//传递到host的/proc，使host的也有问题，在mount隔离下，是不应该有share的mount point的.

	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return errors.WithMessage(err, "PivotRoot.Mount.1.Error")
	}

	/**
	  为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
	  bind mount是把相同的内容换了一个挂载点的挂载方法
	*/
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return errors.WithMessage(err, "PivotRoot.Mount.2.Error")
	}

	// 创建 rootfs/.pivot_root  存储old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("PivotRoot.Mkdir.Error | root=%s | dir=%s", root, pivotDir))
	}

	// pivot_root到新的rootfs，老的old_root现在挂载到rootfs/.pivot_root上
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("PivotRoot.PivotRoot.Error | root=%s | dir=%s", root, pivotDir))
	}
	//修改当前工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("PivotRoot.Chdir.Error | root=%s | dir=%s", root, pivotDir))

	}
	//umount rootfs/.pivot_root
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("PivotRoot.Unmount.Error | root=%s | dir=%s", root, pivotDir))
	}
	// 删除临时目录
	return  os.Remove(pivotDir)
}
