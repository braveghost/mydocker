package container

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/zheng-ji/goSnowFlake"
	"mydocker/setting"
	"os"
	"path"
	"strings"
	"time"
)

var iw *goSnowFlake.IdWorker

func init() {
	 iw, _ = goSnowFlake.NewIdWorker(1)
}
func GetSnowId() string {
	id, _ := iw.NextId()
	return Md5Str(cast.ToString(id))
}

func Md5Str(s string) string {
	hash := md5.New()
	hash.Write([]byte(s))
	value := hash.Sum(nil)
	return hex.EncodeToString(value)
}

type ContainerMeta struct {
	Pid         string `json:"pid"`          // 容器宿主进程id
	Id          string `json:"id"`           // 容器id
	Name        string `json:"name"`         // 容器名称
	Command     string `json:"command"`      // 容器内init进程命令
	CreatedTime string `json:"created_time"` // 创建时间
	Status      string `json:"status"`       // 容器状态
}

func RecordContainerMeta(pid int, cname string, commandArray []string) (string, error) {
	id := GetSnowId()
	ctime := time.Now().Format("2006-01-02 15:04:05")
	if len(cname) == 0{
		cname  = id
	}
	cmeta := &ContainerMeta{
		Pid:         cast.ToString(pid),
		Id:          id,
		Name:        cname,
		Command:     strings.Join(commandArray, " "),
		CreatedTime: ctime,
		Status:      RUNNING,
	}
	jb , err := json.Marshal(cmeta)
	if err != nil {
		logrus.Errorf("RecordContainerMeta.Marshal | %v", err)
		return "", err
	}

	dirs := path.Join(setting.EContainerMetaDataPath, cname)
	if err := os.MkdirAll(dirs,0622); err != nil {
		logrus.Errorf("RecordContainerMeta.MkdirAll | %v", err)
		return "",err
	}
	fname := path.Join(dirs, ConfigName)
	logrus.Info("RecordContainerMeta.Info | %s", fname)
	if f, err := os.Create(fname);err == nil{
		defer f.Close()
		if _, err := f.Write(jb);err != nil{
			logrus.Errorf("RecordContainerMeta.Write | %v", err)
			return "", err
		}
	}else {
		logrus.Errorf("RecordContainerMeta.Create | %v", err)
 	}

	return cname, nil
}

func DeleteContainerMeta(name string)  {
	p := path.Join(setting.EContainerMetaDataPath, name)
	if err := os.RemoveAll(p);err != nil{
		logrus.Errorf("DeleteContainerMeta.RemoveAll | %s | %v", p, err)
	}
}

const(
	RUNNING  = "running"
	STOP  = "stop"
	EXIT  = "exited"

	ConfigName  = "config.json"
)
