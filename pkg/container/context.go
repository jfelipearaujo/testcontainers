package container

import "github.com/testcontainers/testcontainers-go"

// TestContainers is a type that represents a test context
type TestContainers struct {
	Network    *testcontainers.DockerNetwork
	Containers []testcontainers.Container
}

// TestContainersOption is a type that represents a test context option
type TestContainersOption func(*TestContainers)

// WithDockerNetwork is a TestContainersOption that sets the network of the test context
func WithDockerNetwork(network *testcontainers.DockerNetwork) TestContainersOption {
	return func(containers *TestContainers) {
		containers.Network = network
	}
}

// WithDockerContainer is a TestContainersOption that sets the containers of the test context
func WithDockerContainer(container ...testcontainers.Container) TestContainersOption {
	return func(containers *TestContainers) {
		containers.Containers = container
	}
}

func NewTestContainers(opts ...TestContainersOption) TestContainers {
	containers := TestContainers{}

	for _, opt := range opts {
		opt(&containers)
	}

	return containers
}
