package container

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"mydocker/setting"
	"os"
	"os/exec"
	"path"
)

func StopContainer(name string) error {
	meta, err := GetContainerMeta(name)
	if err != nil {
		return err
	}
	pid := meta.Pid

	//syscall.Kill(cast.ToInt(pid), syscall.SIGTERM)
	cmd := exec.Command("kill", "-9", pid)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("StopContainer.Run | %v", err)
		return errors.WithMessage(err, "StopContainer.Run")
	}
	meta.Status = STOP

	data, _ := json.Marshal(meta)
	p := path.Join(setting.EContainerMetaDataPath, name, ConfigName)

	return WriteContainerMeta(p, data)
}
