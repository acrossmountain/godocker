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
	Envs           []string
	ResourceConfig *subsystem.ResourceConfig
}

func newOptions() *Options {
	return &Options{}
}

func (o *Options) apply(opts ...Option) *Options {
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
