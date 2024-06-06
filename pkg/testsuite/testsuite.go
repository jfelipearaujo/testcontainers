package testsuite

import (
	"flag"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var defaultOptions = godog.Options{
	Format:      "pretty",
	Paths:       []string{"features"},
	Output:      colors.Colored(os.Stdout),
	Concurrency: 4,
}

// TestSuiteOption is a type that represents a test suite option
type TestSuiteOption func(*godog.Options)

// WithPaths is a TestSuiteOption that sets the paths of the test suite
//
// Default: "features"
func WithPaths(paths ...string) TestSuiteOption {
	return func(o *godog.Options) {
		o.Paths = paths
	}
}

// WithConcurrency is a TestSuiteOption that sets the concurrency of the test suite
//
// Default: 4
func WithConcurrency(concurrency int) TestSuiteOption {
	return func(o *godog.Options) {
		o.Concurrency = concurrency
	}
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &defaultOptions)
}

// NewTestSuite creates a new test suite
func NewTestSuite(t *testing.T, scenarioInitializer func(ctx *godog.ScenarioContext), opts ...TestSuiteOption) {
	o := defaultOptions
	o.TestingT = t

	for _, opt := range opts {
		opt(&o)
	}

	status := godog.TestSuite{
		ScenarioInitializer: scenarioInitializer,
		Options:             &o,
	}.Run()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}
