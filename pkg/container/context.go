package container

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// GroupContainer is a type that represents a test context
type GroupContainer struct {
	Network    *testcontainers.DockerNetwork
	Containers []testcontainers.Container
}

// TestContainersOption is a type that represents a test context option
type TestContainersOption func(*GroupContainer)

// WithDockerNetwork is a TestContainersOption that sets the network of the test context
func WithDockerNetwork(network *testcontainers.DockerNetwork) TestContainersOption {
	return func(containers *GroupContainer) {
		containers.Network = network
	}
}

// WithDockerContainer is a TestContainersOption that sets the containers of the test context
func WithDockerContainer(container ...testcontainers.Container) TestContainersOption {
	return func(containers *GroupContainer) {
		containers.Containers = container
	}
}

// NewGroup creates a new map of test contexts to store a group of containers
func NewGroup() map[string]GroupContainer {
	return make(map[string]GroupContainer)
}

// BuildGroupContainer creates a new test context with the given options
func BuildGroupContainer(opts ...TestContainersOption) GroupContainer {
	containers := GroupContainer{}

	for _, opt := range opts {
		opt(&containers)
	}

	return containers
}

// DestroyGroup destroys the given group of containers and the network (if exists)
func DestroyGroup(ctx context.Context, group GroupContainer) (context.Context, error) {
	for _, c := range group.Containers {
		err := c.Terminate(ctx)
		if err != nil {
			return ctx, err
		}
	}

	if group.Network != nil {
		if err := group.Network.Remove(ctx); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}
