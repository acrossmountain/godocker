package godocker

import (
	"strconv"

	"godocker/internal/cgroup"
	"godocker/internal/container"
	"godocker/internal/network"

	"github.com/sirupsen/logrus"
)

func Run(tty bool, comArray []string, opts ...container.Option) {
	options := container.NewOptions().Apply(opts...)
	parent, wPipe := container.NewParentProcess(tty, *options)
	if err := parent.Start(); err != nil {
		logrus.Errorf("Start parent procces error: %v", err)
		return
	}

	containerName, err := container.RecordContainerInfo(parent.Process.Pid, options.Name, comArray, options.Volume)
	if err != nil {
		logrus.Errorf("Record container information error: %v", err)
		return
	}

	cGroupManager := cgroup.NewCGroup("godocker.slice")
	defer func() {
		if tty {
			_ = cGroupManager.Destroy()
			// volume imageName containerName
			container.RemoveWorkSpace(options.Volume, options.Name)
			container.RemoveContainerInfo(containerName)
		}
	}()
	_ = cGroupManager.Set(options.ResourceConfig)
	_ = cGroupManager.Apply(parent.Process.Pid)

	if options.Network != "" {
		network.Init()
		containerInfo := &container.Info{
			ID:          containerName,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: options.PortMapping,
		}
		if err := network.ConnectNetwork(options.Network, containerInfo); err != nil {
			logrus.Errorf("Error Connect Network %v", err)
			return
		}
	}

	container.WriteInitCommand(comArray, wPipe)

	if tty {
		_ = parent.Wait()
	}
}
