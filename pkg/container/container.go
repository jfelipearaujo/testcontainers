package container

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/jfelipearaujo/testcontainers/pkg/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// FileData is a type that represents a file data
type FileData struct {
	Reader            io.Reader
	ContainerFilePath string
	FileMode          int64
}

// Container is a type that represents a container
type Container struct {
	Image          string
	ExposedPorts   []string
	EnvVars        map[string]string
	Network        *network.Network
	BaseFilePath   string
	Files          []FileData
	WaitingForLog  string
	StartupTimeout time.Duration
}

// ContainerOption is a type that represents a container option
type ContainerOption func(*Container)

// WithImage is a ContainerOption that sets the image of the container
func WithImage(image string) ContainerOption {
	return func(container *Container) {
		container.Image = image
	}
}

// WithExposedPorts is a ContainerOption that sets the exposed ports of the container
func WithExposedPorts(ports ...string) ContainerOption {
	return func(container *Container) {
		container.ExposedPorts = ports
	}
}

// WithEnvVars is a ContainerOption that sets the environment variables of the container
func WithEnvVars(envVars map[string]string) ContainerOption {
	return func(container *Container) {
		container.EnvVars = envVars
	}
}

// WithNetwork is a ContainerOption that sets the network of the container
func WithNetwork(network *network.Network) ContainerOption {
	return func(container *Container) {
		container.Network = network
	}
}

// WithBaseContainerFilePath is a ContainerOption that sets the base container file path
func WithBaseContainerFilePath(baseContainerFilePath string) ContainerOption {
	return func(container *Container) {
		container.BaseFilePath = baseContainerFilePath
	}
}

// WithFiles is a ContainerOption that sets the startup files of the container
func WithFiles(files ...string) ContainerOption {
	fileData := make([]FileData, len(files))

	for i, file := range files {
		fileData[i] = FileData{
			Reader:            strings.NewReader(file),
			ContainerFilePath: file,
			FileMode:          0644,
		}
	}

	return func(container *Container) {
		container.Files = fileData
	}
}

// WithWaitingForLog is a ContainerOption that sets the log to wait for
func WithWaitingForLog(log string) ContainerOption {
	return func(container *Container) {
		container.WaitingForLog = log
	}
}

// WithStartupTimeout is a ContainerOption that sets the startup timeout
func WithStartupTimeout(timeout time.Duration) ContainerOption {
	return func(container *Container) {
		container.StartupTimeout = timeout
	}
}

// NewContainer creates a new container
func NewContainer(opts ...ContainerOption) *Container {
	container := &Container{}

	for _, opt := range opts {
		opt(container)
	}

	return container
}

// Build creates a new container
func (c *Container) Build(ctx context.Context) (testcontainers.Container, error) {
	var network *testcontainers.DockerNetwork
	var err error

	if c.Network != nil {
		network, err = c.Network.Build(ctx)
		if err != nil {
			return nil, err
		}
	}

	files := make([]testcontainers.ContainerFile, len(c.Files))

	for i, file := range c.Files {
		files[i] = testcontainers.ContainerFile{
			Reader:            file.Reader,
			ContainerFilePath: filepath.Join(c.BaseFilePath, file.ContainerFilePath),
			FileMode:          file.FileMode,
		}
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        c.Image,
			ExposedPorts: c.ExposedPorts,
			Env:          c.EnvVars,
			Networks: []string{
				network.Name,
			},
			NetworkAliases: map[string][]string{
				network.Name: {
					c.Network.Alias,
				},
			},
			Files:      files,
			WaitingFor: wait.ForLog(c.WaitingForLog).WithStartupTimeout(c.StartupTimeout),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create the container: %w", err)
	}
	return container, nil
}
