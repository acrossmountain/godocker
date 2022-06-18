package container

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"godocker/pkg"

	"github.com/sirupsen/logrus"
)

func ListContainer() {
	dir := fmt.Sprintf(RuntimePath, "")
	dir = dir[:len(dir)-1]
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		logrus.Errorf("Read dir %s error %v", dir, err)
		return
	}

	var containers []*Info
	for _, file := range files {
		tmpContainerInfo, err := getContainerInfo(file.Name()) //
		if err != nil {
			logrus.Errorf("Get container info error: %v", err)
			continue
		}
		containers = append(containers, tmpContainerInfo)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.ID,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedAt.String(),
		)
	}

	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error: %v", err)
		return
	}
}

func getContainerInfo(name string) (*Info, error) {
	configFileDir := fmt.Sprintf(RuntimePath, name)
	configFile := path.Join(configFileDir, RuntimeConfigFile)
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		logrus.Errorf("Read file %s erorr %v", configFile, err)
		return nil, err
	}

	var tmpContainer Info
	if err := json.Unmarshal(content, &tmpContainer); err != nil {
		logrus.Errorf("Json unmarshal error: %v", err)
		return nil, err
	}

	return &tmpContainer, nil
}

func getContainerPidByName(name string) (string, error) {
	info, err := getContainerInfo(name)
	if err != nil {
		return "", err
	}

	return info.Pid, nil
}

func RecordContainerInfo(pid int, name string, commands []string, volume string) (string, error) {
	id := pkg.RandStringBytes(10)
	command := strings.Join(commands, "")
	if name == "" {
		name = id
	}

	info := &Info{
		ID:        id,
		Pid:       strconv.Itoa(pid),
		Name:      name,
		CreatedAt: time.Now(),
		Command:   command,
		Status:    Running,
		Volume:    volume,
	}
	buf, err := json.Marshal(info)
	if err != nil {
		logrus.Errorf("Recrod container error: %v", err)
		return "", err
	}

	dir := fmt.Sprintf(RuntimePath, name)
	if err := os.MkdirAll(dir, 0622); err != nil {
		logrus.Errorf("Mkdir  error %s error %v", dir, err)
		return "", err
	}

	fileName := path.Join(dir, RuntimeConfigFile)
	file, err := os.Create(fileName)
	if err != nil {
		logrus.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.WriteString(string(buf)); err != nil {
		logrus.Errorf("File write content error: %v", err)
		return "", err
	}

	return name, nil
}

func RemoveContainerInfo(name string) {
	dir := fmt.Sprintf(RuntimePath, name)
	if err := os.RemoveAll(dir); err != nil {
		logrus.Errorf("Remove dir %s error %v", dir, err)
	}
}
