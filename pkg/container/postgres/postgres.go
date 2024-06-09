package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/jfelipearaujo/testcontainers/pkg/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	BasePath    string = "/docker-entrypoint-initdb.d"
	ExposedPort string = "5432"
	Database    string = "postgres_db"
	User        string = "postgres"
	Pass        string = "postgres"
)

// Options is a type that represents the options for a PostgreSQL container
//
//	Default options:
//		ExposedPort: "5432"
//		Database: "postgres_db"
//		User: "postgres"
//		Pass: "postgres"
//
//	Default network alias: nil
type Options struct {
	ExposedPort  string
	Database     string
	User         string
	Pass         string
	NetworkAlias *string
}

// PostgresOption is a type that represents a PostgreSQL option
type PostgresOption func(*Options)

// WithExposedPort is a PostgresOption that sets the exposed port of the PostgreSQL container
//
//	Default: "5432"
func WithExposedPort(exposedPort string) PostgresOption {
	return func(options *Options) {
		options.ExposedPort = exposedPort
	}
}

// WithDatabase is a PostgresOption that sets the database of the PostgreSQL container
//
//	Default: "postgres_db"
func WithDatabase(database string) PostgresOption {
	return func(options *Options) {
		options.Database = database
	}
}

// WithUser is a PostgresOption that sets the user of the PostgreSQL container
//
//	Default: "postgres"
func WithUser(user string) PostgresOption {
	return func(options *Options) {
		options.User = user
	}
}

// WithPass is a PostgresOption that sets the password of the PostgreSQL container
//
//	Default: "postgres"
func WithPass(pass string) PostgresOption {
	return func(options *Options) {
		options.Pass = pass
	}
}

// WithNetwork is a PostgresOption that sets the network alias of the PostgreSQL container
//
//	Default: nil
func WithNetwork(network *network.Network) PostgresOption {
	return func(options *Options) {
		options.NetworkAlias = &network.Alias
	}
}

// Return a PostgreSQL connection string for the given container with default options
//
//	Example: "postgres://postgres:postgres@localhost:5432/postgres_db?sslmode=disable"
func BuildConnectionString(ctx context.Context, container testcontainers.Container, opts ...PostgresOption) (string, error) {
	var err error

	options := &Options{
		ExposedPort: ExposedPort,
		Database:    Database,
		User:        User,
		Pass:        Pass,
	}

	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get the host: %w", err)
	}

	var mappedPort nat.Port

	mappedPort, err = container.MappedPort(ctx, nat.Port(options.ExposedPort))
	if err != nil {
		return "", fmt.Errorf("failed to get the mapped port: %w", err)
	}

	for _, o := range opts {
		o(options)
	}

	if options.NetworkAlias != nil {
		// changed the host to be the network alias
		host = *options.NetworkAlias

		// changed the mapped port to be the exposed port, allowing the connection to be made between the containers
		mappedPort = nat.Port(options.ExposedPort)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", options.User, options.Pass, host, mappedPort.Port(), options.Database), nil
}

// Return a new container definition for a PostgreSQL container with default options
//
//	DockerImage: "postgres:16"
//	Exposed ports: "5432"
//	Environment variables:
//		POSTGRES_DB: "postgres_db"
//		POSTGRES_USER: "postgres"
//		POSTGRES_PASSWORD: "postgres"
//
//	BasePath: "/docker-entrypoint-initdb.d"
//	WaitingForLog: "database system is ready to accept connections"
//	StartupTimeout: "30 seconds"
func WithPostgresContainer() container.ContainerOption {
	return func(container *container.Container) {
		container.ContainerRequest.Image = "postgres:16"
		container.ContainerRequest.ExposedPorts = []string{
			ExposedPort,
		}
		container.ContainerRequest.Env = map[string]string{
			"POSTGRES_DB":       Database,
			"POSTGRES_USER":     User,
			"POSTGRES_PASSWORD": Pass,
		}
		container.ContainerRequest.WaitingFor = wait.
			ForLog("database system is ready to accept connections").
			WithStartupTimeout(30 * time.Second)
	}
}
