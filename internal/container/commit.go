package container

import (
	"fmt"
	"os/exec"
	"path"

	"github.com/sirupsen/logrus"
)

func CommitContainer(containerName, imageName string) {
	source := fmt.Sprintf(mntPath, containerName) // fmt.Sprintf("%s/%s.tar", rootPath, imageName)
	target := path.Join(rootPath, imageName) + ".tar"
	if _, err := exec.Command("tar", "-czf", target, "-C", source, ".").CombinedOutput(); err != nil {
		logrus.Errorf("Tar folder %s error %v", target, err)
	}
}
