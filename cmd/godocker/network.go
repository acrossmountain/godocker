package godocker

import (
	"fmt"

	"godocker/internal/network"

	"github.com/urfave/cli"
)

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "Manage networks",
	Before: func(ctx *cli.Context) error {
		network.Init()
		return nil
	},
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "Create a network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "Network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "Subnet",
				},
			},
			Action: func(ctx *cli.Context) error {
				if len(ctx.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}

				driver := ctx.String("driver")
				subnet := ctx.String("subnet")
				bridgeName := ctx.Args().First()
				network.CreateNetwork(driver, subnet, bridgeName)
				return nil
			},
		},
		{
			Name:  "rm",
			Usage: "Remove network",
			Action: func(ctx *cli.Context) error {
				if len(ctx.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}

				bridgeName := ctx.Args().First()
				network.DeleteNetwork(bridgeName)
				return nil
			},
		},
		{
			Name:  "ls",
			Usage: "List networks",
			Action: func(ctx *cli.Context) error {
				network.ListNetwork()
				return nil
			},
		},
	},
}

// ip link show dev xxx 查看创建出的 Bridge
// ip addr show dev xxx 查看地址配置和路由配置
// iptables -t nat -vnL POSTROUTING 查看 iptables 配置 MASQUERADE 规则
