package cgroup

import (
	"github.com/pkg/errors"
	"github.com/Sirupsen/logrus"
	"mydocker/subsystems"
	log "github.com/Sirupsen/logrus"

)

type CGroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCGroupManager(path string) *CGroupManager {
	return &CGroupManager{
		Path: path,
	}
}

func (c *CGroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, ss := range subsystems.SubsystemsList {
		if err := ss.Set(c.Path, res); err != nil {
			log.Infof("CGroupManager.Set | %v | %s", res, c.Path)


			return errors.WithMessage(err, "CGroupManager.Set.Error")
		}
	}
	return nil
}

func (c *CGroupManager) Apply(pid int) error {
	for _, ss := range subsystems.SubsystemsList {

		if err := ss.Apply(c.Path, pid); err != nil {
			log.Infof("CGroupManager.Apply | %v | %s", pid,c.Path)

			return errors.WithMessage(err, "CGroupManager.Apply.Error")
		}
	}
	return nil
}
func (c *CGroupManager) Destroy() error {
	for _, ss := range subsystems.SubsystemsList {
		if err := ss.Remove(c.Path); err != nil {
			logrus.Warnf(errors.WithMessage(err, "CGroupManager.Remove.Error").Error())
		}
	}
	return nil
}
