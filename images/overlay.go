package images

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
)

func NewWorkSpaceOverlay(rootUrl string) (string, string) {
	//CreateReadOnlyLayerOverlay(rootUrl)
	upper, work := CreateWriteWorkLayerOverlay(rootUrl)
	CreateMountPointOverlay(rootUrl, upper, work)
	return upper, work
}

func CreateReadOnlyLayerOverlay(rootUrl string) {
	busyboxUrl := rootUrl + "busybox/"
	busybosTarUrl := rootUrl + "busybox.tar"
	exist := PathExists(busyboxUrl)

	if !exist {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			logrus.Errorf("CreateReadOnlyLayerOverlay.Mkdir | %v", err)
		}
		if _, err := exec.Command("tar", "-xvf", busybosTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
			logrus.Errorf("CreateReadOnlyLayerOverlay.Command | %v", err)
		}
	}
}
func GetWriteWorkLayerOverlay(rootUrl string) (string, string) {

	return rootUrl + "_upperdir", rootUrl + "_workdir"
}

func CreateWriteWorkLayerOverlay(rootUrl string) (string, string) {
	upperDir, workDir := GetWriteWorkLayerOverlay(rootUrl)

	if exist := PathExists(upperDir); exist {
		logrus.Errorf("CreateMountPointOverlay.PathExists.UpperDir | %s", upperDir)
	} else {
		if err := os.Mkdir(upperDir, 0777); err != nil {
			logrus.Errorf("CreateMountPointOverlay.Mkdir.UpperDir | %v | %s", err, upperDir)
		}
	}

	if exist := PathExists(workDir); exist {
		logrus.Errorf("CreateMountPointOverlay.PathExists.WorkDir | %v | %s", workDir)
	} else {
		if err := os.Mkdir(workDir, 0777); err != nil {
			logrus.Errorf("CreateMountPointOverlay.Mkdir.WorkDir | %v | %s", err, workDir)
		}

	}
	return upperDir, workDir
}
func CreateMountPointOverlay(rootUrl, upperUrl, workUrl string) {

	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", rootUrl, upperUrl, workUrl)

	//mount -t overlay overlay -o lowerdir=lower1:lower2,upperdir=upper,workdir=work merged
	logrus.Infof("CreateMountPointOverlay.Path | %s", dirs)

	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, workUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("CreateMountPointOverlay.Run | %s", err.Error())

	}
}
