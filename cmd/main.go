package main

import (
	"godocker/cmd/godocker"

	"github.com/sirupsen/logrus"
)

func main() {
	app := godocker.NewApp()
	if err := app.Run(); err != nil {
		logrus.Error(err)
	}
}
