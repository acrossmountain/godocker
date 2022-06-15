package godocker

import (
	"fmt"
	"os"

	"godocker/internal/cgroup/subsystem"
	"godocker/internal/container"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var containerCommand = cli.Command{
	Name:  "container",
	Usage: "Manage containers",
	Subcommands: []cli.Command{
		runCommand,
		execCommand,
		commitCommand,
		listCommand,
		logsCommand,
		removeCommand,
	},
}

// sudo ./godocker run -it -m 100m "stress --vm-types 200m --vm-keep -m 1"
var runCommand = cli.Command{
	Name:  "run",
	Usage: "Run a command in a new container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "Run container in background",
		},
		cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "Assign a name to the container",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "Set environment variables",
		},
	},
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}

		tty := ctx.Bool("it")
		volume := ctx.String("v")
		detach := ctx.Bool("d")
		name := ctx.String("name")

		if tty && detach {
			return fmt.Errorf("it and d paramter can't both provided")
		}

		res := &subsystem.ResourceConfig{
			MemoryLimit: ctx.String("mem"),
			CpuShare:    ctx.String("cpushare"),
			CpuSet:      ctx.String("cpuset"),
		}
		commands := ctx.Args()
		image := commands[0]
		commands = commands[1:]
		envs := ctx.StringSlice("e")

		container.Run(tty, commands,
			container.WithContainerName(name),
			container.WithResourceConfig(res),
			container.WithVolume(volume),
			container.WithDetach(detach),
			container.WithImage(image),
			container.WithEnv(envs),
		)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container",
	Action: func(ctx *cli.Context) error {
		return container.RunInitProcess()
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "Run a command in a running container",
	Action: func(ctx *cli.Context) error {
		if os.Getenv(container.ENV_EXEC_PID) != "" {
			logrus.Infof("pid callbackp pid %d", os.Getgid())
			return nil
		}

		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container name or command")
		}
		containerName := ctx.Args().First()
		cmdArray := ctx.Args().Tail()

		container.Exec(containerName, cmdArray)
		return nil
	},
}

var listCommand = cli.Command{
	Name:  "ls",
	Usage: "List containers",
	Action: func(ctx *cli.Context) error {
		container.ListContainer()
		return nil
	},
}

var logsCommand = cli.Command{
	Name:  "logs",
	Usage: "Fetch the logs of a container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("please input yout container name")
		}
		containerName := ctx.Args().First()
		container.LogContainer(containerName)
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "Create a new image from a container's changes",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		containerName := ctx.Args().Get(0)
		imageName := ctx.Args().Get(1)

		container.CommitContainer(containerName, imageName)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "Stop running container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}
		containerName := ctx.Args().First()
		container.StopContainer(containerName)

		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "Remove container",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing container command")
		}

		containerName := ctx.Args().First()
		container.RemoveContainer(containerName)

		return nil
	},
}
