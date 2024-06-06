package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/testcontainers/testcontainers-go"
)

const (
	Port     string = "5432"
	Database string = "postgres_db"
	User     string = "postgres"
	Pass     string = "postgres"
)

func ConnectionString(ctx context.Context, container testcontainers.Container) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get the host: %w", err)
	}

	port, err := container.MappedPort(ctx, nat.Port(Port))
	if err != nil {
		return "", fmt.Errorf("failed to get the mapped port: %w", err)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", User, Pass, host, port, Database), nil
}

// Create a Postgres container with the default configuration
func WithPostgresContainer() container.ContainerOption {
	return func(container *container.Container) {
		container.Image = "postgres:latest"
		container.ExposedPorts = []string{
			Port,
		}
		container.EnvVars = map[string]string{
			"POSTGRES_DB":       Database,
			"POSTGRES_USER":     User,
			"POSTGRES_PASSWORD": Pass,
		}
		container.BaseFilePath = "/docker-entrypoint-initdb.d"
		container.WaitingForLog = "ready for start up"
		container.StartupTimeout = 30 * time.Second
	}
}
