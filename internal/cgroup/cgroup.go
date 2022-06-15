package cgroup

import (
	"godocker/internal/cgroup/subsystem"

	"github.com/sirupsen/logrus"
)

type CGroup struct {
	Path           string
	ResourceConfig *subsystem.ResourceConfig
}

func NewCGroup(path string) *CGroup {
	return &CGroup{
		Path: path,
	}
}

func (c *CGroup) Apply(pid int) error {
	for _, subSys := range subsystem.SubSystems() {
		if err := subSys.Apply(c.Path, pid); err != nil {
			logrus.Errorf("CGroup apply %s error %v", subSys.Name(), err)
		}
	}

	return nil
}

func (c *CGroup) Set(resConfig *subsystem.ResourceConfig) error {
	for _, subSys := range subsystem.SubSystems() {
		if err := subSys.Set(c.Path, resConfig); err != nil {
			logrus.Errorf("CGroup set  %s error %v", subSys.Name(), err)
		}
	}

	return nil
}

func (c *CGroup) Destroy() error {
	for _, subSys := range subsystem.SubSystems() {
		if err := subSys.Remove(c.Path); err != nil {
			logrus.Errorf("CGroup remove  %s error %v", subSys.Name(), err)
		}
	}

	return nil
}
