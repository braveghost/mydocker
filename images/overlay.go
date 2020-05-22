package images

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"mydocker/setting"
	"os"
	"os/exec"
	"path"
)

func NewWorkSpaceOverlay(image string) (string, string) {
	rootUrl := CreateReadOnlyLayerOverlay(image)
	upper, work, ok := CreateWriteWorkLayerOverlay()
	if ok {
		CreateMountPointOverlay(rootUrl, upper, work)
	}
	return upper, work
}

func CreateReadOnlyLayerOverlay(image string) string {
	busyboxUrl := path.Join(setting.EContainerPath, image)
	busybosTarUrl := path.Join(setting.EImagesPath, image+".tar")
	exist := PathExists(busyboxUrl)

	if !exist {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			logrus.Errorf("CreateReadOnlyLayerOverlay.Mkdir | %v", err)
		}
		if _, err := exec.Command("tar", "-xvf", busybosTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
			logrus.Errorf("CreateReadOnlyLayerOverlay.Command | %v", err)
		}
	}

	return busyboxUrl
}
func GetWriteWorkLayerOverlay() (string, string) {

	return setting.EContainerPath + "_upperdir", setting.EContainerPath + "_workdir"
}

func CreateWriteWorkLayerOverlay() (string, string, bool) {
	upperDir, workDir := GetWriteWorkLayerOverlay()
	var ok = true
	if exist := PathExists(upperDir); exist {
		ok = false

		logrus.Errorf("CreateMountPointOverlay.PathExists.UpperDir | %s", upperDir)
	} else {
		if err := os.Mkdir(upperDir, 0777); err != nil {
			logrus.Errorf("CreateMountPointOverlay.Mkdir.UpperDir | %v | %s", err, upperDir)
		}
	}

	if exist := PathExists(workDir); exist {
		ok = false
		logrus.Errorf("CreateMountPointOverlay.PathExists.WorkDir | %s", workDir)
	} else {
		if err := os.Mkdir(workDir, 0777); err != nil {
			logrus.Errorf("CreateMountPointOverlay.Mkdir.WorkDir | %v | %s", err, workDir)
		}

	}
	return upperDir, workDir, ok
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
