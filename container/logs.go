package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/pkg/errors"
	"io/ioutil"
	"mydocker/setting"
	"os"
	"path"
)

func LogsContainer(name string) error {
	logFile := path.Join(setting.EContainerLogsDataPath, name, setting.EContainerLogName)

	f, err := os.Open(logFile)
	defer f.Close()
	if err != nil {
		logrus.Errorf("LogsContainer.Open | %v", err)
		return errors.WithMessage(err, "LogsContainer.Open")
	}
	c, err := ioutil.ReadAll(f)
	if err != nil {
		logrus.Errorf("LogsContainer.ReadAll | %v", err)
		return errors.WithMessage(err, "LogsContainer.ReadAll")
	}
	fmt.Fprint(os.Stdout, string(c))
	return nil
}
