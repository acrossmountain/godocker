package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type Status string

const (
	Running Status = "running"
	Stop    Status = "stop"
	Exit    Status = "exit"
)

type Info struct {
	ID          string    `json:"id"`
	Pid         string    `json:"pid"`
	Name        string    `json:"name"`
	Command     string    `json:"command"`
	Status      Status    `json:"status"`
	Volume      string    `json:"volume"`
	PortMapping []string  `json:"port_mapping"`
	CreatedAt   time.Time `json:"created_at"`
}

var (
	selfProcessExe    = "/proc/self/exe"
	rootPath          = "/home/kexin/projects/godocker/"
	mntPath           = "/home/kexin/projects/godocker/mnt/%s"
	wirtePath         = "/home/kexin/projects/godocker/write/%s"
	RuntimePath       = "/var/run/godocker/%s"
	RuntimeConfigFile = "config.json"
	RuntimeLogFile    = "container.log"
)

// 1. /proc/self/exe 调用中，/proc/self/ 指的是当前运行进程自己的环境，exec 其实就是调用了自己，使用这种方式对自己进行初始化。
// 2. args 是参数，其中 init 是传递给本进程的第一个参数。
// 3. clone 参数就是 namespace 隔离标识。
// 4. 如果用户指定了 -it 参数，就需要把当前的输出、输入、错误导入到表主输出上
func NewParentProcess(tty bool, options Options) (*exec.Cmd, *os.File) {
	r, w, err := os.Pipe()
	if err != nil {
		logrus.Errorf("New pipe error: %v", err)
		return nil, nil
	}
	cmd := exec.Command(selfProcessExe, "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNS,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		dir := fmt.Sprintf(RuntimePath, options.Name)
		if err := os.MkdirAll(dir, 0622); err != nil {
			logrus.Errorf("NewParentProcess mkdir %s error %v", dir, err)
			return nil, nil
		}

		stdLogFile := path.Join(dir, RuntimeLogFile)
		file, err := os.Create(stdLogFile)
		if err != nil {
			logrus.Errorf("NewParentProcess create file %s error %v", stdLogFile, err)
			return nil, nil
		}

		cmd.Stdout = file
		cmd.Stderr = file
	}

	// volume imageName containerName
	NewWorkSpace(options.Volume, options.Image, options.Name)

	cmd.ExtraFiles = []*os.File{r}
	cmd.Dir = fmt.Sprintf(mntPath, options.Name)    // 进程启动时的目录.
	cmd.Env = append(os.Environ(), options.Envs...) // env

	return cmd, w
}

func RunInitProcess() error {
	cmdArray := readInitCommand()
	if len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}

	setUpMount()

	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		logrus.Errorf("Exec loop path error: %v", err)
		return err
	}
	logrus.Infof("Find path %s", path)
	// 执行当前 filename 对应程序。覆盖当前进程的镜像、数据和堆栈等信息，包括PID。
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		logrus.Error(err)
	}

	return nil
}

func StopContainer(name string) {
	pid, err := getContainerPidByName(name)
	if err != nil {
		logrus.Errorf("Get container pid by name %s error %v", name, err)
		return
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logrus.Errorf("Conver pid from striing to int error %v", err)
		return
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		logrus.Errorf("Stop container %s error %v", name, err)
		return
	}

	containerInfo, err := getContainerInfo(name)
	if err != nil {
		logrus.Errorf("Get container %s info %v", name, err)
		return
	}

	containerInfo.Pid = ""
	containerInfo.Status = Stop

	content, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("Json marshal %s error %v", name, err)
		return
	}

	configPath := path.Join(fmt.Sprintf(RuntimePath, name), RuntimeConfigFile)
	if err := ioutil.WriteFile(configPath, content, 0622); err != nil {
		logrus.Errorf("Write file %s error %v", configPath, err)
	}
}

func RemoveContainer(name string) {
	// remove container
	containerInfo, err := getContainerInfo(name)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", name, err)
		return
	}

	if containerInfo.Status != Stop {
		logrus.Errorf("Couldn't remove running container.")
		return
	}

	dir := fmt.Sprintf(RuntimePath, name)
	if err := os.RemoveAll(dir); err != nil {
		logrus.Errorf("Remove file %s error %v", dir, err)
		return
	}

	// todo: remove mount path and write path.
}
