package container

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"io/ioutil"
	"mydocker/setting"
	"mydocker/utils"
	"os"
	"path"
	"strings"
	"time"
)

type ContainerMeta struct {
	Pid         string   `json:"pid"`          // 容器宿主进程id
	Id          string   `json:"id"`           // 容器id
	Name        string   `json:"name"`         // 容器名称
	Command     string   `json:"command"`      // 容器内init进程命令
	CreatedTime string   `json:"created_time"` // 创建时间
	Status      string   `json:"status"`       // 容器状态
	Image       string   `json:"image"`        // 镜像
	Volume      string   `json:"volume"`       // 卷
	PortMapping []string `json:"port_mapping"` // 端口
	Ip          string   `json:"ip"`           // ip
}

func NewContainerMeta(pid int, cname, image, volume string, commandArray, portMapping []string) *ContainerMeta {
	id := utils.GetSnowId()
	if len(cname) == 0 {
		cname = id
	}
	ctime := time.Now().Format("2006-01-02 15:04:05")

	return &ContainerMeta{
		Pid:         cast.ToString(pid),
		Id:          id,
		Name:        cname,
		Command:     strings.Join(commandArray, " "),
		CreatedTime: ctime,
		Status:      RUNNING,
		Image:       image,
		PortMapping: portMapping,
		Volume:      volume,
	}
}

func RecordContainerMeta(meta *ContainerMeta) error {

	jb, err := json.Marshal(meta)
	if err != nil {
		logrus.Errorf("RecordContainerMeta.Marshal | %v", err)
		return err
	}

	dirs := path.Join(setting.EContainerMetaDataPath, meta.Name)
	if err := os.MkdirAll(dirs, 0622); err != nil {
		logrus.Errorf("RecordContainerMeta.MkdirAll | %v", err)
		return err
	}
	fname := path.Join(dirs, ConfigName)
	logrus.Info("RecordContainerMeta.Info | %s", fname)
	if f, err := os.Create(fname); err == nil {
		defer f.Close()
		if _, err := f.Write(jb); err != nil {
			logrus.Errorf("RecordContainerMeta.Write | %v", err)
			return err
		}
	} else {
		logrus.Errorf("RecordContainerMeta.Create | %v", err)
	}
	return nil
}

func GetMetaPath(name string) string {
	return path.Join(setting.EContainerMetaDataPath, name)
}

// 更新元数据
func WriteContainerMeta(meta *ContainerMeta) error {

	data, _ := json.Marshal(meta)
	p := path.Join(setting.EContainerMetaDataPath, meta.Name, ConfigName)

	if err := ioutil.WriteFile(p, data, 0622); err != nil {
		logrus.Errorf("WriteContainerMeta.WriteFile | %v", err)
		return errors.WithMessage(err, "WriteContainerMeta.WriteFile")
	}
	return nil
}

func DeleteContainerMeta(name string) {
	metaPath := GetMetaPath(name)
	if err := os.RemoveAll(metaPath); err != nil {
		logrus.Errorf("DeleteContainerMeta.RemoveAll | %s | %v", metaPath, err)
	}

}

func GetContainerMeta(name string) (*ContainerMeta, error) {
	p := path.Join(setting.EContainerMetaDataPath, name, ConfigName)
	c, err := ioutil.ReadFile(p)
	if err != nil {
		logrus.Errorf("GetContainerMeta.ReadFile | %v", err)
		return nil, errors.WithMessage(err, "GetContainerMeta.ReadFile ")
	}
	meta := &ContainerMeta{}
	if err := json.Unmarshal(c, meta); err != nil {
		logrus.Errorf("GetContainerMeta.Unmarshal | %v", err)

		return nil, errors.WithMessage(err, "GetContainerMeta.Unmarshal")
	}
	return meta, nil
}

func UpdateContainerStatus(name, status string) error {
	meta, err := GetContainerMeta(name)
	if err != nil {
		logrus.Errorf("UpdateContainerStatus.GetContainerMeta | %v", err)
		return errors.WithMessage(err, "UpdateContainerStatus.GetContainerMeta")
	}
	meta.Status = STOP
	return WriteContainerMeta(meta)
}

const (
	RUNNING = "running"
	STOP    = "stop"
	EXIT    = "exited"

	ConfigName = "config.json"
)
