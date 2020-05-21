package images

import (
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
)

func NewWorkSpaceAufs(rootUrl, mntUrl string) {
	CreateReadOnlyLayerAufs(rootUrl)
	CreateWriteLayerAufs(rootUrl)
	CreateMountPointAufs(rootUrl, mntUrl)
}

func CreateReadOnlyLayerAufs(rootUrl string) {
	busyboxUrl := rootUrl + "busybox/"
	busybosTarUrl := rootUrl + "busybox.tar"
	exist:= PathExists(busyboxUrl)

	if !exist {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			logrus.Errorf("CreateReadOnlyLayerAufs.Mkdir | %v", err)
		}
		if _, err := exec.Command("tar", "-xvf", busybosTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
			logrus.Errorf("CreateReadOnlyLayerAufs.Command | %v", err)
		}
	}
}
func CreateWriteLayerAufs(rootUrl string) {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.Mkdir(writeUrl, 0777); err != nil {
		logrus.Errorf("CreateWriteLayerAufs.Mkdir | %v", err)
	}

}
func CreateMountPointAufs(rootUrl, mntUrl string) {
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		logrus.Errorf("CreateMountPointAufs.Mkdir | %v", err)
	}
	dirs := "dirs=" + rootUrl + "writeLayer:" + rootUrl + "busybox"
	cmd := exec.Command("mount", "-t", "overlay", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("CreateMountPointAufs.Run | %s", err.Error())

	}
}

func PathExists(url string) bool {
	_, err := os.Stat(url)
	if err == nil {
		return true
	}
	return false
}
