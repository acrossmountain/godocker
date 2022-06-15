package subsystem

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

func findCGroupMountPoint(subSys string) string {
	// /proc/self/mountinfo 当前进程的挂载点信息
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	// 41 32 0:36 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:19 - cgroup cgroup rw,memory
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ") // xxx xx xx memory
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subSys {
				return fields[4] // 第五列
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ""
	}

	return ""
}

func getCGroupPath(subSys string, cGroupPath string, authCreate bool) (string, error) {
	cGroupRoot := findCGroupMountPoint(subSys)
	_, err := os.Stat(path.Join(cGroupRoot, cGroupPath))
	if err == nil || (authCreate && os.IsNotExist(err)) { // 如果等
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cGroupRoot, cGroupPath), os.FileMode(0755)); err == nil {
			} else {
				return "", fmt.Errorf("error create cgroup %v", err)
			}
		}
		return path.Join(cGroupRoot, cGroupPath), nil
	}

	return "", fmt.Errorf("cgroup path error %v", err)
}
