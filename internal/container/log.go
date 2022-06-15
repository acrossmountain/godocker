package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

func LogContainer(name string) {
	logFilePath := path.Join(fmt.Sprintf(RuntimePath, name), RuntimeLogFile)
	file, err := os.Open(logFilePath)
	if err != nil {
		logrus.Errorf("Log container open file %s error %v", logFilePath, err)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("Log container read file %s error %v", logFilePath, err)
		return
	}

	_, _ = fmt.Fprintf(os.Stdout, string(content))
}


