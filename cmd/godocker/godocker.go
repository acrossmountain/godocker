package godocker

import (
	"os"

	_ "godocker/internal/nsenter"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Docker struct {
	cliApp *cli.App
}

func NewApp() *Docker {
	cliApp := cli.NewApp()
	cliApp.Usage = "godocker is a simple container runtime"
	cliApp.Before = func(ctx *cli.Context) error {
		logrus.SetReportCaller(true)
		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(os.Stdout)

		return nil
	}
	cliApp.Commands = []cli.Command{
		initCommand,
		runCommand,
		execCommand,
		commitCommand,
		listCommand,
		logsCommand,
		stopCommand,
		removeCommand,
		containerCommand,
	}

	return &Docker{cliApp: cliApp}
}

func (app *Docker) Run() error {
	return app.cliApp.Run(os.Args)
}
