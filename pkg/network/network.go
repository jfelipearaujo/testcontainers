package network

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

// NetworkType is a type that represents the type of network
type NetworkType string

var (
	// NetworkTypeBridge is a network type that represents a bridge network
	NetworkTypeBridge NetworkType = "bridge"
)

// Network is a type that represents a network
type Network struct {
	Alias string
	Type  NetworkType

	instance *testcontainers.DockerNetwork
}

// NetworkOption is a type that represents a network option
type NetworkOption func(*Network)

// WithAlias is a NetworkOption that sets the alias of the network
//
// Default: network
func WithAlias(alias string) NetworkOption {
	return func(network *Network) {
		network.Alias = alias
	}
}

// WithType is a NetworkOption that sets the type of the network
//
// Default: bridge
func WithType(typeName NetworkType) NetworkOption {
	return func(network *Network) {
		network.Type = typeName
	}
}

// NewNetwork creates a new Network
func NewNetwork(opts ...NetworkOption) *Network {
	ntw := &Network{
		Alias: "network",
		Type:  NetworkTypeBridge,
	}

	for _, opt := range opts {
		opt(ntw)
	}

	return ntw
}

// Build creates a new DockerNetwork
func (ntw *Network) Build(ctx context.Context) (*testcontainers.DockerNetwork, error) {
	if ntw.instance != nil {
		return ntw.instance, nil
	}

	output, err := tcnetwork.New(ctx, tcnetwork.WithDriver(string(ntw.Type)))
	if err != nil {
		return nil, fmt.Errorf("failed to create the network: %w", err)
	}

	ntw.instance = output

	return output, nil
}
