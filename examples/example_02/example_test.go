package example

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/jfelipearaujo/testcontainers/pkg/container"
	"github.com/jfelipearaujo/testcontainers/pkg/container/postgres"
	"github.com/jfelipearaujo/testcontainers/pkg/state"
	"github.com/jfelipearaujo/testcontainers/pkg/testsuite"

	_ "github.com/lib/pq"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type user struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type test struct {
	connStr   string
	userName  string
	userEmail string
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
			postgres.WithPostgresContainer(),
			container.WithFiles(postgres.BasePath, "./testdata/init.sql"),
		)

		pgContainer, err := definition.BuildContainer(ctx)
		if err != nil {
			return ctx, err
		}

		connectionString, err := postgres.BuildConnectionString(ctx, pgContainer)
		if err != nil {
			return ctx, err
		}

		currentState := testState.Retrieve(ctx)
		currentState.connStr = connectionString

		containers[sc.Id] = container.BuildGroupContainer(
			container.WithDockerContainer(pgContainer),
		)

		return testState.Enrich(ctx, currentState), nil
	})

	ctx.Step(`^I have entered "([^"]*)" into the user name field$`, iHaveEnteredIntoTheUserNameField)
	ctx.Step(`^I have entered "([^"]*)" into the user email field$`, iHaveEnteredIntoTheUserEmailField)
	ctx.Step(`^I press "([^"]*)"$`, iPress)
	ctx.Step(`^the user should be created$`, theUserShouldBeCreated)
	ctx.Step(`^the user should be updated$`, theUserShouldBeUpdated)
	ctx.Step(`^the user should be deleted$`, theUserShouldBeDeleted)

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		group := containers[sc.Id]

		return container.DestroyGroup(ctx, group)
	})
}

func iHaveEnteredIntoTheUserNameField(ctx context.Context, name string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.userName = name
	return testState.Enrich(ctx, currentState), nil
}

func iHaveEnteredIntoTheUserEmailField(ctx context.Context, email string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)
	currentState.userEmail = email
	return testState.Enrich(ctx, currentState), nil
}

func iPress(ctx context.Context, operation string) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	operations := map[string]func(ctx context.Context) error{
		"create": func(ctx context.Context) error {
			return createUser(currentState.connStr,
				currentState.userName,
				currentState.userEmail)
		},
		"update": func(ctx context.Context) error {
			return updateUser(currentState.connStr,
				currentState.userName,
				currentState.userEmail)
		},
		"delete": func(ctx context.Context) error {
			return deleteUser(currentState.connStr,
				currentState.userEmail)
		},
	}

	operationHandler, ok := operations[operation]
	if !ok {
		return ctx, fmt.Errorf("operation '%s' not supported", operation)
	}

	if err := operationHandler(ctx); err != nil {
		return ctx, fmt.Errorf("failed to '%s' user: %w", operation, err)
	}

	return ctx, nil
}

func theUserShouldBeCreated(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	_, err := selectUser(currentState.connStr,
		currentState.userName,
		currentState.userEmail)
	if err != nil {
		return ctx, fmt.Errorf("failed to read user: %w", err)
	}

	return ctx, nil
}

func theUserShouldBeUpdated(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	user, err := selectUser(currentState.connStr,
		currentState.userName,
		currentState.userEmail)
	if err != nil {
		return ctx, fmt.Errorf("failed to read user: %w", err)
	}

	if user.Name != currentState.userName {
		return ctx, fmt.Errorf("user name not updated")
	}

	return ctx, nil
}

func theUserShouldBeDeleted(ctx context.Context) (context.Context, error) {
	currentState := testState.Retrieve(ctx)

	_, err := selectUser(currentState.connStr,
		currentState.userName,
		currentState.userEmail)
	if err == ErrUserNotFound {
		return ctx, nil
	}
	if err != nil {
		return ctx, fmt.Errorf("failed to read user: %w", err)
	}

	return ctx, nil
}

func createUser(connStr, name, email string) error {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	//defer conn.Close()

	_, err = conn.Exec(
		"INSERT INTO users (name, email) VALUES ($1, $2);",
		name,
		email)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func selectUser(connStr, name, email string) (user, error) {
	user := user{}

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return user, fmt.Errorf("failed to open connection: %w", err)
	}
	//defer conn.Close()

	rows, err := conn.Query(
		"SELECT * FROM users WHERE name = $1 AND email = $2",
		name,
		email)
	if err != nil {
		return user, fmt.Errorf("failed to read user: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return user, ErrUserNotFound
	}

	err = rows.Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		return user, fmt.Errorf("failed to scan user: %w", err)
	}

	return user, nil
}

func updateUser(connStr, name, email string) error {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	//defer conn.Close()

	result, err := conn.Exec("UPDATE users SET name = $1 WHERE email = $2",
		name,
		email)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affectedRows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func deleteUser(connStr, email string) error {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	//defer conn.Close()

	result, err := conn.Exec("DELETE FROM users WHERE email = $1",
		email)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affectedRows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
