package container

import (
	"github.com/sirupsen/logrus"
	"mydocker/images"
	"os"
	"syscall"
)

func RemoveContainer(name string) error {
	meta, err := GetContainerMeta(name)
	if err != nil {
		return err
	}
	if meta.Status != STOP{
		logrus.Errorf("RemoveContainer.NotStop | %s", name)
		return nil
	}

	upperPath, workPath := images.GetWriteWorkLayerOverlay(name)

	if err := syscall.Unmount(workPath, 0); err != nil {
		logrus.Errorf("RemoveContainer.Unmount | %s | %s", workPath, err.Error())
	}

	if err := os.RemoveAll(upperPath); err != nil {
		logrus.Errorf("RemoveContainer.RemoveAll.UpperPath | %s | %s", upperPath, err.Error())
	}

	if err := os.RemoveAll(workPath); err != nil {
		logrus.Errorf("RemoveContainer.RemoveAll.WorkPath | %s | %s", workPath, err.Error())
	}
	logPath := GetLogPath(name)
	if err := os.RemoveAll(logPath); err != nil {
		logrus.Errorf("RemoveContainer.RemoveAll.Log | %s | %s", logPath, err.Error())
	}
	metaPath := GetMetaPath(name)
	if err := os.RemoveAll(metaPath); err != nil {
		logrus.Errorf("RemoveContainer.RemoveAll.Meta | %s | %s", metaPath, err.Error())
	}

	return nil

}
