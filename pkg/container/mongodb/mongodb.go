package mongodb

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
	ExposedPort string = "27017"
	User        string = "mongo"
	Pass        string = "mongo"
)

// Options is a type that represents the options for a MongoDB container
//
//	Default options:
//		ExposedPort: "27017"
//		User: "mongo"
//		Pass: "mongo"
type Options struct {
	ExposedPort string
	User        string
	Pass        string
}

// MongoOption is a type that represents a MongoDB option
type MongoOption func(*Options)

// WithExposedPort is a MongoOption that sets the exposed port of the MongoDB container
//
//	Default: "27017"
func WithExposedPort(exposedPort string) MongoOption {
	return func(options *Options) {
		options.ExposedPort = exposedPort
	}
}

// WithUser is a MongoOption that sets the user of the MongoDB container
//
//	Default: "test"
func WithUser(user string) MongoOption {
	return func(options *Options) {
		options.User = user
	}
}

// WithPass is a MongoOption that sets the password of the MongoDB container
//
//	Default: "test"
func WithPass(pass string) MongoOption {
	return func(options *Options) {
		options.Pass = pass
	}
}

// Return a MongoDB connection string for the given container with default options
//
//	Example: "mongodb://mongo:mongo@localhost:27017/"
func BuildConnectionString(ctx context.Context, container testcontainers.Container, opts ...MongoOption) (string, error) {
	options := &Options{
		ExposedPort: ExposedPort,
		User:        User,
		Pass:        Pass,
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

	return fmt.Sprintf("mongodb://%s:%s@%s:%s/", options.User, options.Pass, host, port.Port()), nil
}

// Return a new container definition for a MongoDB container with default options
//
//	DockerImage: "mongo:7"
//	Exposed ports: "27017"
//	Environment variables:
//		MONGO_INITDB_ROOT_USERNAME: "mongo"
//		MONGO_INITDB_ROOT_PASSWORD: "mongo"
//
//	WaitingForLog: "Waiting for connections"
//	StartupTimeout: "30 seconds"
func WithMongoContainer() container.ContainerOption {
	return func(container *container.Container) {
		container.ContainerRequest.Image = "mongo:7"
		container.ContainerRequest.ExposedPorts = []string{
			ExposedPort,
		}
		container.ContainerRequest.Env = map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": User,
			"MONGO_INITDB_ROOT_PASSWORD": Pass,
		}
		container.ContainerRequest.WaitingFor = wait.
			ForLog("Waiting for connections").
			WithStartupTimeout(30 * time.Second)
	}
}
