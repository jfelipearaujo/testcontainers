package example

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
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

func initializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^I have entered "([^"]*)" into the calculator$`, iHaveEnteredIntoTheCalculator)
	ctx.Step(`^I press "([^"]*)"$`, iPress)
	ctx.Step(`^the result should be "([^"]*)" on the screen$`, theResultShouldBeOnTheScreen)
}

func iHaveEnteredIntoTheCalculator(ctx context.Context, number float64) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.numbers = append(currentState.numbers, number)
	return testState.Enrich(ctx, currentState), nil
}

func iPress(ctx context.Context, operation string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if len(currentState.numbers) != 2 {
		return ctx, fmt.Errorf("invalid number of numbers: %d, expected 2", len(currentState.numbers))
	}

	operations := map[string]func(a float64, b float64) float64{
		"add": func(a float64, b float64) float64 {
			return a + b
		},
		"subtract": func(a float64, b float64) float64 {
			return a - b
		},
		"multiply": func(a float64, b float64) float64 {
			return a * b
		},
		"divide": func(a float64, b float64) float64 {
			return a / b
		},
	}

	calc, ok := operations[operation]
	if !ok {
		return ctx, fmt.Errorf("operation not found: %s", operation)
	}

	currentState.result = calc(currentState.numbers[0], currentState.numbers[1])

	return testState.Enrich(ctx, currentState), nil
}

func theResultShouldBeOnTheScreen(ctx context.Context, expected float64) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	if currentState.result != expected {
		return ctx, fmt.Errorf("result is not correct: %f, expected %f", currentState.result, expected)
	}

	return ctx, nil
}
