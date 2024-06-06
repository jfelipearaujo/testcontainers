package example

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/jfelipearaujo/testcontainers/pkg/container/postgres"
	"github.com/jfelipearaujo/testcontainers/pkg/network"
	"github.com/jfelipearaujo/testcontainers/pkg/state"
	"github.com/jfelipearaujo/testcontainers/pkg/testsuite"
)

func TestFeatures(t *testing.T) {
	testsuite.NewTestSuite(t,
		initializeScenario,
		testsuite.WithPaths("features"),
		testsuite.WithConcurrency(4),
	)
}

type test struct {
	numbers []float64
	result  float64
	connStr string
}

var testState = state.NewState[test]()

var (
	containers = make(map[string]container.TestContainers)
)

func initializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		network := network.NewNetwork()
		definition := container.NewContainer(
			container.WithNetwork(network),
			postgres.WithPostgresContainer(),
			container.WithFiles("./testdata/init.sql"),
		)

		pgContainer, err := definition.Build(ctx)
		if err != nil {
			return ctx, err
		}

		connectionString, err := postgres.ConnectionString(ctx, pgContainer)
		if err != nil {
			return ctx, err
		}

		currentState := testState.Retrieve(ctx)
		currentState.connStr = connectionString

		containers[sc.Id] = container.NewTestContainers(
			container.WithDockerNetwork(network.GetInstance()),
			container.WithDockerContainer(pgContainer),
		)

		return testState.Enrich(ctx, currentState), nil
	})

	ctx.Step(`^I have entered "([^"]*)" into the user name field$`, iHaveEnteredIntoTheUserNameField)
	ctx.Step(`^I have entered "([^"]*)" into the user email field$`, iHaveEnteredIntoTheUserEmailField)
	ctx.Step(`^I have an existing user$`, iHaveAnExistingUser)
	ctx.Step(`^I press "([^"]*)"$`, iPress)
	ctx.Step(`^the user should be created$`, theUserShouldBeCreated)
	ctx.Step(`^the user should be read$`, theUserShouldBeRead)
	ctx.Step(`^the user should be updated$`, theUserShouldBeUpdated)
	ctx.Step(`^the user should be deleted$`, theUserShouldBeDeleted)

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if err != nil {
			return ctx, err
		}

		tc := containers[sc.Id]

		for _, c := range tc.Containers {
			err := c.Terminate(ctx)
			if err != nil {
				return ctx, err
			}
		}

		if tc.Network != nil {
			if err := tc.Network.Remove(ctx); err != nil {
				return ctx, err
			}
		}

		return ctx, nil
	})
}

func iHaveEnteredIntoTheUserNameField(ctx context.Context, name string) (context.Context, error) {
	return ctx, nil
}

func iHaveEnteredIntoTheUserEmailField(ctx context.Context, email string) (context.Context, error) {
	return ctx, nil
}

func iHaveAnExistingUser(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func iPress(ctx context.Context, operation string) (context.Context, error) {
	return ctx, nil
}

func theUserShouldBeCreated(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func theUserShouldBeRead(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func theUserShouldBeUpdated(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func theUserShouldBeDeleted(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
