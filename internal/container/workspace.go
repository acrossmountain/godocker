package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"godocker/pkg"

	"github.com/sirupsen/logrus"
)

func NewWorkSpace(volume, imageName, containerName string) {
	createReadOnlyLayer(imageName)
	createWriteLayer(containerName)
	createMountPoint(containerName, imageName)

	if volume != "" {
		volumes := volumeUrlExtract(volume)
		if len(volumes) == 2 && volumes[0] != "" && volumes[1] != "" {
			mountVolume(volumes, containerName)
		} else {
			logrus.Infof("Volume parameter input is not correct.")
		}
	}
}

// createReadOnlyLayer Create container readonly layer
// 1. mkdir container dir.
// 2. unTar image to container readonly dir
func createReadOnlyLayer(imageName string) {
	target := path.Join(rootPath, imageName)
	source := fmt.Sprintf("%s/%s.tar", rootPath, imageName)

	exist, err := pkg.PathExists(target)
	if err != nil {
		logrus.Infof("Fail to judge whether dir %s exists. %v", target, err)
	}

	if !exist {
		if err := os.Mkdir(target, 0755); err != nil {
			logrus.Errorf("Mkdir dir %s error. %v", target, err)
		}
	}

	if _, err := exec.Command("tar", "-xvf", source, "-C", target).CombinedOutput(); err != nil {
		logrus.Errorf("Untar dir %s error %v", source, target)
	}
}

// createWriteLayer Create container write layer.
func createWriteLayer(containerName string) {
	target := fmt.Sprintf(wirtePath, containerName)
	if err := os.Mkdir(target, 0777); err != nil {
		logrus.Infof("Mkdir writelayer dir %s error: %v", target, err)
	}
}

// createMountPoint
func createMountPoint(containerName, imageName string) {
	if err := os.Mkdir(fmt.Sprintf(mntPath, containerName), 0777); err != nil {
		logrus.Infof("Mkdir mount path: %s, error: %v", mntPath, err)
	}

	// fmt.Sprintf("dirs=%writeLayer:%sbusybox", rootPath, rootPath)
	dirs := "dirs=" + fmt.Sprintf(wirtePath, containerName) + ":" + path.Join(rootPath, imageName)
	logrus.Infof("only dirs: %s, mnt dir: %s", dirs, fmt.Sprintf(mntPath, containerName))
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", fmt.Sprintf(mntPath, containerName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Mount point error: %v", err)
	}
}

// volume imageName containerName
func RemoveWorkSpace(volume, containerName string) {
	if volume != "" {
		volumes := volumeUrlExtract(volume)
		if len(volumes) == 2 && volumes[0] != "" && volumes[1] != "" {
			umountVolume(volumes, containerName)
		}
	}

	removeMountPoint(containerName)
	// removeReadOnlyLayer(rootPath)
	removeWriteLayer(containerName)
}

func removeReadOnlyLayer(imageName string) {
	dir := path.Join(rootPath, "busybox")
	if err := os.RemoveAll(dir); err != nil {
		logrus.Infof("Remove busybox dir %s error: %v", dir, err)
	}
}

func removeWriteLayer(containerName string) {
	dir := fmt.Sprintf(wirtePath, containerName) // path.Join(rootPath, "writeLayer")
	if err := os.RemoveAll(dir); err != nil {
		logrus.Infof("Remove writelayer dir %s error: %v", dir, err)
	}
}

func removeMountPoint(containerName string) {
	cmd := exec.Command("umount", fmt.Sprintf(mntPath, containerName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("Umount dir %s error %v", fmt.Sprintf(mntPath, containerName), err)
	}
	if err := os.RemoveAll(fmt.Sprintf(mntPath, containerName)); err != nil {
		logrus.Infof("Remove mount dir %s error: %v", fmt.Sprintf(mntPath, containerName), err)
	}
}
