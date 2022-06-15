package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

func volumeUrlExtract(volume string) []string {
	return strings.Split(volume, ":")
}

func mountVolume(volumes []string, containerName string) {
	// 创建宿主机目录
	hostPath := volumes[0]
	if err := os.MkdirAll(hostPath, 0777); err != nil {
		logrus.Infof("Mkdir host volume dir: %s, error: %v", hostPath, err)
	}

	// 在容器系统里创建挂载点
	containerVolumePath := path.Join(fmt.Sprintf(mntPath, containerName), volumes[1])
	if err := os.Mkdir(containerVolumePath, 0777); err != nil {
		logrus.Infof("Mkdir container volume dir: %s, error: %v", containerVolumePath, err)
	}

	// 宿主机文件目录挂载到容器挂载点
	dirs := "dirs=" + hostPath
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Mount volume %s failed. %v", containerVolumePath, err)
	}
}

func umountVolume(volumes []string, containerName string) {
	containerVolumePath := path.Join(mntPath, containerName, volumes[1])
	cmd := exec.Command("umount", containerVolumePath)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Umount container volume %s failed. %v", containerVolumePath, err)
	}
}
