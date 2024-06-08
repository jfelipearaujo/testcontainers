package localstack

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	BasePath    string = "/etc/localstack/init/ready.d"
	ExposedPort string = "4566"
	Debug       string = "false"
	DockerHost  string = "unix:///var/run/docker.sock"
)

// Options is a type that represents the options for a LocalStack container
//
//	Default options:
//		ExposedPort: "4566"
//		Debug: false
//		DockerHost: "unix:///var/run/docker.sock"
type Options struct {
	ExposedPort string
	Debug       string
	DockerHost  string
}

// LocalStackOption is a type that represents a LocalStack option
type LocalStackOption func(*Options)

// WithExposedPort is a LocalStackOption that sets the exposed port of the LocalStack container
//
//	Default: "4566"
func WithExposedPort(exposedPort string) LocalStackOption {
	return func(options *Options) {
		options.ExposedPort = exposedPort
	}
}

// WithDebug is a LocalStackOption that sets the debug of the LocalStack container
//
//	Default: false
func WithDebug(debug string) LocalStackOption {
	return func(options *Options) {
		options.Debug = debug
	}
}

// WithDockerHost is a LocalStackOption that sets the Docker host of the LocalStack container
//
// Default: unix:///var/run/docker.sock
func WithDockerHost(dockerHost string) LocalStackOption {
	return func(options *Options) {
		options.DockerHost = dockerHost
	}
}

// BuildEndpoint returns the endpoint of the LocalStack container
//
//	Example: "http://localhost:4566"
func BuildEndpoint(ctx context.Context, container testcontainers.Container, opts ...LocalStackOption) (string, error) {
	options := &Options{
		ExposedPort: ExposedPort,
		Debug:       Debug,
		DockerHost:  DockerHost,
	}

	for _, o := range opts {
		o(options)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get the host: %w", err)
	}

	port, err := container.MappedPort(ctx, nat.Port(options.ExposedPort))
	if err != nil {
		return "", fmt.Errorf("failed to get the mapped port: %w", err)
	}

	return fmt.Sprintf("http://%s:%s", host, port.Port()), nil
}

// Return a new container definition for a LocalStack container with default options:
//
//	DockerImage: "localstack/localstack:3.4"
//	ExposedPort: "4566"
//
//	Environment variables:
//		DEBUG: false
//		DOCKER_HOST: "unix:///var/run/docker.sock"
//
//	BasePath: "/etc/localstack/init/ready.d"
//	WaitingForLog: "Initialization complete!"
//	StartupTimeout: "30 seconds"
func WithLocalStackContainer() container.ContainerOption {
	return func(container *container.Container) {
		container.ContainerRequest.Image = "localstack/localstack:3.4"
		container.ContainerRequest.ExposedPorts = []string{
			ExposedPort,
		}
		container.ContainerRequest.Env = map[string]string{
			"DEBUG":       Debug,
			"DOCKER_HOST": DockerHost,
		}
		container.ContainerRequest.WaitingFor = wait.
			ForLog("Initialization complete!").
			WithStartupTimeout(30 * time.Second)
	}
}
