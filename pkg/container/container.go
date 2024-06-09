package container

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Container is a type that represents a container that will be created
type Container struct {
	ContainerRequest  testcontainers.ContainerRequest
	ForceWaitDuration *time.Duration
}

// ContainerOption is a type that represents a container option
type ContainerOption func(*Container)

// WithDockerfile is a ContainerOption that sets the Dockerfile data of the container
//
// Default: nil
func WithDockerfile(fromDockerFile testcontainers.FromDockerfile) ContainerOption {
	return func(container *Container) {
		container.ContainerRequest.FromDockerfile = fromDockerFile
	}
}

// WithImage is a ContainerOption that sets the image of the container
//
// Default: postgres:latest
func WithImage(image string) ContainerOption {
	return func(container *Container) {
		container.ContainerRequest.Image = image
	}
}

// WithExposedPorts is a ContainerOption that sets the exposed ports of the container
//
// Default: 5432
func WithExposedPorts(ports ...string) ContainerOption {
	return func(container *Container) {
		container.ContainerRequest.ExposedPorts = ports
	}
}

// WithEnvVars is a ContainerOption that sets the environment variables of the container
//
// Default:
//
//	POSTGRES_DB: postgres_db
//	POSTGRES_USER: postgres
//	POSTGRES_PASSWORD: postgres
func WithEnvVars(envVars map[string]string) ContainerOption {
	return func(container *Container) {
		container.ContainerRequest.Env = envVars
	}
}

// WithNetwork is a ContainerOption that sets the network of the container
//
// Default: nil
func WithNetwork(alias string, network *testcontainers.DockerNetwork) ContainerOption {
	return func(container *Container) {
		container.ContainerRequest.Networks = []string{
			network.Name,
		}
		container.ContainerRequest.NetworkAliases = map[string][]string{
			network.Name: {
				alias,
			},
		}
	}
}

// WithFiles is a ContainerOption that sets the startup files of the container that will be copied to the container
//
// Default: nil
func WithFiles(basePath string, files ...string) ContainerOption {
	fileData := make([]testcontainers.ContainerFile, len(files))

	if len(files) == 0 {
		panic(fmt.Errorf("files must not be empty"))
	}

	for i, file := range files {
		reader, err := os.Open(file)
		if err != nil {
			panic(fmt.Errorf("failed to open file '%s': %w", file, err))
		}
		fileData[i] = testcontainers.ContainerFile{
			Reader:            reader,
			ContainerFilePath: filepath.Join(basePath, filepath.Base(file)),
			FileMode:          0644,
		}
	}

	return func(container *Container) {
		container.ContainerRequest.Files = fileData
	}
}

// WithExecutableFiles is a ContainerOption that sets the executable files of the container that will be copied to the container
//
// Default: nil
func WithExecutableFiles(basePath string, files ...string) ContainerOption {
	fileData := make([]testcontainers.ContainerFile, len(files))

	if len(files) == 0 {
		panic(fmt.Errorf("executable files must not be empty"))
	}

	for i, file := range files {
		reader, err := os.Open(file)
		if err != nil {
			panic(fmt.Errorf("failed to open file '%s': %w", file, err))
		}
		fileData[i] = testcontainers.ContainerFile{
			Reader:            reader,
			ContainerFilePath: filepath.Join(basePath, filepath.Base(file)),
			FileMode:          0755,
		}
	}

	return func(container *Container) {
		container.ContainerRequest.Files = fileData
	}
}

// WithWaitingForLog is a ContainerOption that sets the log to wait for
//
// Default: ready for start up
func WithWaitingForLog(log string, startupTimeout time.Duration) ContainerOption {
	return func(container *Container) {
		container.ContainerRequest.WaitingFor = wait.ForLog(log).WithStartupTimeout(startupTimeout)
	}
}

// WithWaitingForPort is a ContainerOption that sets the port to wait for
//
//	Example: "8080" for 30 seconds
func WithWaitingForPort(port string, startupTimeout time.Duration) ContainerOption {
	return func(container *Container) {
		container.ContainerRequest.WaitingFor = wait.ForListeningPort(nat.Port(port)).WithStartupTimeout(startupTimeout)
	}
}

// WithForceWaitDuration is a ContainerOption that sets the duration to wait for the container to be ready
//
//	Default: nil
func WithForceWaitDuration(duration time.Duration) ContainerOption {
	return func(container *Container) {
		container.ForceWaitDuration = &duration
	}
}

// NewContainerDefinition creates a new container definition that will be used to create a container
func NewContainerDefinition(opts ...ContainerOption) *Container {
	container := &Container{
		ContainerRequest: testcontainers.ContainerRequest{},
	}

	for _, opt := range opts {
		opt(container)
	}

	return container
}

// BuildContainer creates a new container following the container definition
func (c *Container) BuildContainer(ctx context.Context) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: c.ContainerRequest,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create the container: %w", err)
	}

	if c.ForceWaitDuration != nil {
		fmt.Printf("Waiting for %s for the container to be ready\n", *c.ForceWaitDuration)
		time.Sleep(*c.ForceWaitDuration)
	}

	return container, nil
}

func GetMappedPort(ctx context.Context, container testcontainers.Container, exposedPort nat.Port) (nat.Port, error) {
	containerID := container.GetContainerID()

	cmd := exec.Command("docker", "inspect", "--format", "{{json .NetworkSettings.Ports}}", containerID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute docker inspect: %s, %v", output, err)
	}

	var ports nat.PortMap
	if err = json.Unmarshal(output, &ports); err != nil {
		return "", fmt.Errorf("failed to parse output: %v", err)
	}

	port, ok := ports[exposedPort]
	if !ok {
		return "", fmt.Errorf("port %s not found", exposedPort)
	}

	if len(port) == 0 {
		return "", fmt.Errorf("port %s not found", exposedPort)
	}

	return nat.Port(port[0].HostPort), nil
}
