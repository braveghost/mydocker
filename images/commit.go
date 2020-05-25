package images

import (
	"github.com/sirupsen/logrus"
	"mydocker/setting"
	"os/exec"
	"path"
)

func CommitImage(name string)  {
	_,work := GetWriteWorkLayerOverlay(name)
	imageUrl := path.Join(setting.EImagesPath, name + ".tar")
	logrus.Infof("CommitImage.Path | %s", imageUrl)
  	if _, err := exec.Command("tar", "-czf", imageUrl, "-C", work,".").CombinedOutput();err != nil{
		logrus.Errorf("CommitImage.Command | %v", err)
	}

}
