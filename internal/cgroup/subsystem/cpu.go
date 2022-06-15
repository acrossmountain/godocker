package subsystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuSubSys struct {
}

func (s *CpuSubSys) Name() string {
	return "cpu"
}

func (s *CpuSubSys) Set(cGroupPath string, res *ResourceConfig) error {
	subSysCgroupPath, err := getCGroupPath(s.Name(), cGroupPath, true)
	if err != nil {
		return err
	}

	if res.CpuShare != "" {
		if err := ioutil.WriteFile(path.Join(subSysCgroupPath, "cpu.shares"), []byte(res.CpuShare), 0644); err != nil {
			return fmt.Errorf("set cgroup memory fail %v", err)
		}
	}

	return nil
}

func (s *CpuSubSys) Apply(cGroupPath string, pid int) error {
	subSysCgroupPath, err := getCGroupPath(s.Name(), cGroupPath, false)
	if err != nil {
		return fmt.Errorf("get cgroup %s error %v", cGroupPath, err)
	}

	err = ioutil.WriteFile(path.Join(subSysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}

	return nil
}

func (s *CpuSubSys) Remove(cGroupPath string) error {
	subSysCgroupPath, err := getCGroupPath(s.Name(), cGroupPath, false)
	if err == nil {
		return os.RemoveAll(subSysCgroupPath)
	}

	return nil
}
