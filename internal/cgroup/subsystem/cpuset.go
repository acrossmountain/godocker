package subsystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuSetSubSys struct {
}

func (s *CpuSetSubSys) Name() string {
	return "cpuset"
}

func (s *CpuSetSubSys) Set(cGroupPath string, res *ResourceConfig) error {
	subSysCgroupPath, err := getCGroupPath(s.Name(), cGroupPath, true)
	if err != nil {
		return err
	}

	if res.CpuSet != "" {
		if err := ioutil.WriteFile(path.Join(subSysCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644); err != nil {
			return fmt.Errorf("set cgroup memory fail %v", err)
		}
	}

	return nil
}

func (s *CpuSetSubSys) Apply(cGroupPath string, pid int) error {
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

func (s *CpuSetSubSys) Remove(cGroupPath string) error {
	subSysCgroupPath, err := getCGroupPath(s.Name(), cGroupPath, false)
	if err == nil {
		return os.RemoveAll(subSysCgroupPath)
	}

	return nil
}

