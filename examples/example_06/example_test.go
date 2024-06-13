package example

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/go-connections/nat"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/jfelipearaujo/testcontainers/pkg/container/postgres"
	"github.com/jfelipearaujo/testcontainers/pkg/network"
	"github.com/jfelipearaujo/testcontainers/pkg/state"
	"github.com/jfelipearaujo/testcontainers/pkg/testsuite"
	"github.com/testcontainers/testcontainers-go"
)

type product struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type test struct {
	apiUrl string

	productId   int
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
		ntwrkDefinition := network.NewNetwork()

		network, err := ntwrkDefinition.Build(ctx)
		if err != nil {
			return ctx, fmt.Errorf("failed to build the network: %w", err)
		}

		pgDefinition := container.NewContainerDefinition(
			container.WithNetwork(ntwrkDefinition.Alias, network),
			postgres.WithPostgresContainer(),
			container.WithFiles(postgres.BasePath, "./testdata/init.sql"),
		)

		pgContainer, err := pgDefinition.BuildContainer(ctx)
		if err != nil {
			return ctx, err
		}

		connString, err := postgres.BuildInternalConnectionString(ctx, pgContainer, postgres.WithNetwork(ntwrkDefinition))
		if err != nil {
			return ctx, err
		}

		apiDefinition := container.NewContainerDefinition(
			container.WithNetwork(ntwrkDefinition.Alias, network),
			container.WithDockerfile(testcontainers.FromDockerfile{
				Context:    "./custom_api",
				Dockerfile: "Dockerfile",
				KeepImage:  false,
			}),
			container.WithEnvVars(map[string]string{
				"DATABASE_URL": connString,
			}),
			container.WithExposedPorts("8080"),
			container.WithWaitingForLog("server running on port 8080", 10*time.Second),
		)

		apiContainer, err := apiDefinition.BuildContainer(ctx)
		if err != nil {
			return ctx, err
		}

		host, err := apiContainer.Host(ctx)
		if err != nil {
			return ctx, fmt.Errorf("failed to get the host: %w", err)
		}

		port, err := container.GetMappedPort(ctx, apiContainer, nat.Port("8080/tcp"))
		if err != nil {
			return ctx, err
		}

		currentState := testState.Retrieve(ctx)
		currentState.apiUrl = fmt.Sprintf("http://%s:%s", host, port)

		containers[sc.Id] = container.BuildGroupContainer(
			container.WithDockerContainer(pgContainer),
			container.WithDockerContainer(apiContainer),
		)

		return testState.Enrich(ctx, currentState), nil
	})

	ctx.Step(`^I have a product with the id "([^"]*)"$`, iHaveAProductWithTheId)
	ctx.Step(`^I have a product with the name "([^"]*)"$`, iHaveAProductWithTheName)
	ctx.Step(`^I create the product$`, iCreateTheProduct)
	ctx.Step(`^I retrieve the product with the id "([^"]*)"$`, iRetrieveTheProductWithTheId)
	ctx.Step(`^the product should be created$`, theProductShouldBeCreated)
	ctx.Step(`^the product should be retrieved$`, theProductShouldBeRetrieved)
	ctx.Step(`^the product should not be retrieved$`, theProductShouldNotBeRetrieved)

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		group := containers[sc.Id]

		return container.DestroyGroup(ctx, group)
	})
}

func iHaveAProductWithTheId(ctx context.Context, productId int) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.productId = productId
	return testState.Enrich(ctx, currentState), nil
}

func iHaveAProductWithTheName(ctx context.Context, productName string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.productName = productName
	return testState.Enrich(ctx, currentState), nil
}

func iCreateTheProduct(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	reqBody := `{
		"id": ` + strconv.Itoa(currentState.productId) + `,
		"name": "` + currentState.productName + `"
	}`

	client := &http.Client{}
	route := fmt.Sprintf("%s/products", currentState.apiUrl)
	req, err := http.NewRequest("POST", route, strings.NewReader(reqBody))
	if err != nil {
		return ctx, fmt.Errorf("failed to create the request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to send the request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return ctx, fmt.Errorf("failed to create the product: %d", resp.StatusCode)
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

func iRetrieveTheProductWithTheId(ctx context.Context, productId int) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	client := &http.Client{}
	route := fmt.Sprintf("%s/products/%d", currentState.apiUrl, productId)
	req, err := http.NewRequest("GET", route, nil)
	if err != nil {
		return ctx, fmt.Errorf("failed to create the request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to send the request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// this is intentional, we don't want to fail the test if the product doesn't exist
		currentState.product = product{}
		return testState.Enrich(ctx, currentState), nil
	}

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

func theProductShouldBeCreated(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if currentState.product.Id != currentState.productId {
		return ctx, fmt.Errorf("product not created, expected id %d, got %d", currentState.productId, currentState.product.Id)
	}

	if currentState.product.Name != currentState.productName {
		return ctx, fmt.Errorf("product not created, expected name %s, got %s", currentState.productName, currentState.product.Name)
	}

	return ctx, nil
}

func theProductShouldBeRetrieved(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if currentState.product.Id != currentState.productId {
		return ctx, fmt.Errorf("product not created, expected id %d, got %d", currentState.productId, currentState.product.Id)
	}

	if currentState.product.Name != currentState.productName {
		return ctx, fmt.Errorf("product not created, expected name %s, got %s", currentState.productName, currentState.product.Name)
	}

	return ctx, nil
}

func theProductShouldNotBeRetrieved(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if currentState.product.Id != 0 {
		return ctx, fmt.Errorf("product created, expected id 0, got %d", currentState.product.Id)
	}

	if currentState.product.Name != "" {
		return ctx, fmt.Errorf("product created, expected name \"\", got %s", currentState.product.Name)
	}

	return ctx, nil
}
