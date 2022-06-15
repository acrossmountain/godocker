package subsystem

var subSystems []SubSystem

func init() {
	subSystems = []SubSystem{
		&MemorySubSys{},
		&CpuSubSys{},
		&CpuSetSubSys{},
	}
}

func SubSystems() []SubSystem {
	return subSystems
}

type ResourceConfig struct {
	MemoryLimit string	// 内存限制
	CpuShare    string  // CPU 时间片权重
	CpuSet      string  // CPU 核心数
}

type SubSystem interface {
	// Name subsystem 对象名称
	Name() string

	// Set 设置某个 cgroup 在这个 subsystem 中限制
	Set(cGroupPath string, res *ResourceConfig) error

	// Apply 将进程添加到某个 cgroup 中
	Apply(cGroupPath string, pid int) error

	// Remove 设置 cgroup
	Remove(cGroupPath string) error
}
