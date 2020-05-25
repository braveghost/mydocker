package container

import (
	"github.com/sirupsen/logrus"
	"mydocker/images"
	"os"
	"syscall"
)

func RemoveContainer() error {
	upperPath, workPath := images.GetWriteWorkLayerOverlay()

	if err := syscall.Unmount(workPath,0);err != nil{
		logrus.Errorf("RemoveContainer.Unmount | %s | %s", workPath, err.Error())
	}

	if err := os.RemoveAll(upperPath); err != nil {
		logrus.Errorf("RemoveContainer.RemoveAll.UpperPath | %s | %s", upperPath, err.Error())
	}

	if err := os.RemoveAll(workPath); err != nil {
		logrus.Errorf("RemoveContainer.RemoveAll.WorkPath | %s | %s", workPath, err.Error())
	}

	return nil

}
