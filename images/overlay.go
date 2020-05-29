package images

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"mydocker/setting"
	"os"
	"os/exec"
	"path"
)

/*
创建容器的写入层
@params image: 镜像名称
@params name: 容器名称
@return string: 更新层
@return string: 工作层
*/
func NewWorkSpaceOverlay(image string, name string) (string, string) {
	// 通过镜像创建容器的只读层
	rootUrl := CreateReadOnlyLayerOverlay(image)
	// 创建容器的写入层
	upper, work := CreateWriteWorkLayerOverlay(name)
	// 挂载更新层和工作层
	CreateMountPointOverlay(rootUrl, upper, work)
	return upper, work
}

/*
通过镜像创建容器的只读层
@params image: 颈项名称
@return string: 容器只读层路径
*/
func CreateReadOnlyLayerOverlay(image string) string {
	busyboxUrl := path.Join(setting.EContainerPath, image)
	busybosTarUrl := path.Join(setting.EImagesPath, image+".tar")
	exist := PathExists(busyboxUrl)

	if !exist {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			logrus.Infof("CreateReadOnlyLayerOverlay.Mkdir | %v", err)
		}
		if _, err := exec.Command("tar", "-xvf", busybosTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
			logrus.Errorf("CreateReadOnlyLayerOverlay.Command | %v | %s", err, busybosTarUrl)
			// 没有容器直接退出
			os.Exit(1)
		}
	}

	return busyboxUrl
}

/*
构建更新层目录和工作目录
@params name: 容器名称
@return string: 更新层目录路径
@return string: 工作层目录路径
*/
func GetWriteWorkLayerOverlay(name string) (string, string) {
	return path.Join(setting.EContainerPath+"_upperdir", name), path.Join(setting.EContainerPath+"_workdir", name)
}

/*
创建容器的写入层
@params name: 容器名称
@return string: 更新层路径
@return string: 工作层路径
@return bool: 工作层
*/
func CreateWriteWorkLayerOverlay(name string) (string, string) {
	upperDir, workDir := GetWriteWorkLayerOverlay(name)

	if err := os.Mkdir(upperDir, 0777); err != nil {
		logrus.Errorf("CreateMountPointOverlay.Mkdir.UpperDir | %v | %s", err, upperDir)
	}

	if err := os.Mkdir(workDir, 0777); err != nil {
		logrus.Errorf("CreateMountPointOverlay.Mkdir.WorkDir | %v | %s", err, workDir)
	}
	return upperDir, workDir
}

/*
挂载更新层和工作层
@params rootUrl: 只读层路径
@return upperUrl: 更新层路径
@return workUrl: 工作层路径
*/
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
