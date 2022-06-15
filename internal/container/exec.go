package container

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

//
const (
	ENV_EXEC_PID = "go_docker_pid"
	ENV_EXEC_CMD = "go_docker_cmd"
)

func Exec(containerName string, comArray []string) {
	// runtime.LockOSThread()
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("Exec container getCOntainerPidByName %s error %v", containerName, err)
		return
	}

	cmdStr := strings.Join(comArray, " ")
	logrus.Infof("PID %d, Command %s", pid, cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Exec container %s error %v", containerName, err)
	}

}
