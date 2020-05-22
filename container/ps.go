package container

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"io/ioutil"
	"mydocker/setting"
	"os"
	"path"
	"text/tabwriter"
)

func ListContainer() error {
	s, err := os.Stat(setting.EContainerMetaDataPath)
	if err != nil {
		logrus.Errorf("ListContainer.Stat | %v", err)
		return errors.WithMessage(err, "ListContainer.Stat")
	}
	if s.IsDir() {
		files, err := ioutil.ReadDir(setting.EContainerMetaDataPath)
		if err != nil {
			logrus.Errorf("ListContainer.ReadDir | %v", err)
			return errors.WithMessage(err, "ListContainer.ReadDir")
		}
		var containers []*ContainerMeta
		for _, f := range files {
			tmp, err := getContainerInfo(f)
			if err != nil {
				return errors.WithMessage(err, "ListContainer.GetContainerInfo")
			}
			containers = append(containers, tmp)
		}
		// 输出格式化
		w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
		fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
		for _, i := range containers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", i.Id, i.Name, i.Pid, i.Status,
				i.Command, i.CreatedTime)
		}
		if err := w.Flush(); err != nil {
			logrus.Errorf("ListContainer.Flush | %v", err)
			return errors.WithMessage(err, "ListContainer.Flush")
		}
	}
	return nil
}

// 获取所有元信息文件
func getContainerInfo(file os.FileInfo) (*ContainerMeta, error) {
	// 容器id
	name := file.Name()
	// 元信息文件路径
	cfd := path.Join(setting.EContainerMetaDataPath, name, ConfigName)
	// 读取信息
	content, err := ioutil.ReadFile(cfd)
	if err != nil {
		logrus.Errorf("ListContainer.GetContainerInfo.ReadDir | %v | %s", err, cfd)
		return nil, err
	}
	c := &ContainerMeta{}
	if err := json.Unmarshal(content, c); err != nil {
		logrus.Errorf("ListContainer.GetContainerInfo.Unmarshal | %v | %s", err, cfd)
		return nil, err
	}
	return c, nil
}
