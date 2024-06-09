package example

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/go-connections/nat"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/jfelipearaujo/testcontainers/pkg/state"
	"github.com/jfelipearaujo/testcontainers/pkg/testsuite"
	"github.com/testcontainers/testcontainers-go"
)

type product struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type test struct {
	apiUrl string

	productName string

	product product
}

var testState = state.NewState[test]()

var containers = container.NewGroup()

func TestFeatures(t *testing.T) {
	testsuite.NewTestSuite(t,
		initializeScenario,
		testsuite.WithPaths("features"),
		testsuite.WithConcurrency(0),
	)
}

func initializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		definition := container.NewContainerDefinition(
			container.WithDockerfile(testcontainers.FromDockerfile{
				Context:    "./custom_api",
				Dockerfile: "Dockerfile",
				KeepImage:  false,
			}),
			container.WithExposedPorts("8080"),
			container.WithWaitingForLog("server running on port 8080", 10*time.Second),
		)

		apiContainer, err := definition.BuildContainer(ctx)
		if err != nil {
			return ctx, err
		}

		host, err := apiContainer.Host(ctx)
		if err != nil {
			return ctx, fmt.Errorf("failed to get the host: %w", err)
		}

		ports, err := apiContainer.MappedPort(ctx, nat.Port("8080"))
		if err != nil {
			return ctx, fmt.Errorf("failed to get the mapped port: %w", err)
		}

		currentState := testState.Retrieve(ctx)
		currentState.apiUrl = fmt.Sprintf("http://%s:%s", host, ports.Port())

		containers[sc.Id] = container.BuildGroupContainer(
			container.WithDockerContainer(apiContainer),
		)

		return testState.Enrich(ctx, currentState), nil
	})

	ctx.Step(`^I have a product with name "([^"]*)"$`, iHaveAProductWithName)
	ctx.Step(`^I get the product$`, iGetTheProduct)
	ctx.Step(`^the product should be returned$`, theProductShouldBeReturned)

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		group := containers[sc.Id]

		return container.DestroyGroup(ctx, group)
	})
}

func iHaveAProductWithName(ctx context.Context, productName string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.productName = productName
	return testState.Enrich(ctx, currentState), nil
}

func iGetTheProduct(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	client := &http.Client{}
	route := fmt.Sprintf("%s/products", currentState.apiUrl)
	req, err := http.NewRequest("GET", route, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to create the request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to send the request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ctx, fmt.Errorf("failed to get the product: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx, fmt.Errorf("failed to read the response body: %w", err)
	}

	var product product

	if err := json.Unmarshal(body, &product); err != nil {
		return ctx, fmt.Errorf("failed to unmarshal the response body: %w", err)
	}

	currentState.product = product

	return testState.Enrich(ctx, currentState), nil
}

func theProductShouldBeReturned(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if currentState.product.Name != currentState.productName {
		return ctx, fmt.Errorf("product not returned")
	}

	return ctx, nil
}
