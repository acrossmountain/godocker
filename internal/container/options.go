package container

import (
	"fmt"

	"godocker/internal/cgroup/subsystem"
)

type Option func(opts *Options)

type Options struct {
	Name           string
	Image          string
	TTY            bool
	Detach         bool
	Volume         string
	Network        string
	Envs           []string
	PortMapping    []string
	ResourceConfig *subsystem.ResourceConfig
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Apply(opts ...Option) *Options {
	for _, opt := range opts {
		opt(o)
	}

	fmt.Println(o)
	return o
}

func WithResourceConfig(resourceConfig *subsystem.ResourceConfig) Option {
	return func(opts *Options) {
		opts.ResourceConfig = resourceConfig
	}
}

func WithVolume(volume string) Option {
	return func(opts *Options) {
		opts.Volume = volume
	}
}

func WithDetach(detach bool) Option {
	return func(opts *Options) {
		opts.Detach = detach
	}
}

func WithTTY(tty bool) Option {
	return func(opts *Options) {
		opts.TTY = tty
	}
}

func WithContainerName(containerName string) Option {
	return func(opts *Options) {
		opts.Name = containerName
	}
}

func WithImage(image string) Option {
	return func(opts *Options) {
		opts.Image = image
	}
}

func WithEnv(envs []string) Option {
	return func(opts *Options) {
		opts.Envs = envs
	}
}

func WithNetwork(network string) Option {
	return func(opts *Options) {
		opts.Network = network
	}
}

func WithPortMapping(portMapping []string) Option {
	return func(opts *Options) {
		opts.PortMapping = portMapping
	}
}
