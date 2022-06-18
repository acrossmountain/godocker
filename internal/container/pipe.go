package container

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func WriteInitCommand(comArray []string, w *os.File) {
	command := strings.Join(comArray, " ")
	_, _ = w.WriteString(command)
	_ = w.Close()
}

func readInitCommand() []string {
	f := os.NewFile(uintptr(3), "pipe")
	content, err := ioutil.ReadAll(f)
	if err != nil {
		logrus.Errorf("Init read command error: %v", err)
	}

	return strings.Split(string(content), " ")
}
