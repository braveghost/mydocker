package container

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
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
	}
	meta.Status = STOP

	return WriteContainerMeta(meta)
}
