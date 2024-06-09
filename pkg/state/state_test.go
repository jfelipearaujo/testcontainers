package state_test

import (
	"context"
	"testing"

	"github.com/jfelipearaujo/testcontainers/pkg/state"
	"github.com/stretchr/testify/assert"
)

type MyData struct {
	Val int
}

func TestState(t *testing.T) {
	t.Run("Should be able to retrieve data", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		myData := MyData{
			Val: 1,
		}

		state := state.NewState[MyData]()

		// Act
		ctx = state.Enrich(ctx, &myData)

		// Assert
		res := state.Retrieve(ctx)
		assert.Equal(t, myData.Val, res.Val)
	})

	t.Run("Should return nil if data is not found", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		state := state.NewState[MyData]()

		// Act
		res := state.Retrieve(ctx)

		// Assert
		assert.Empty(t, res)
	})

	t.Run("Should be able to retrieve data with custom key", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		myData := MyData{
			Val: 1,
		}

		state := state.NewState(
			state.WithCtxKey[MyData]("my-key"),
		)

		// Act
		ctx = state.Enrich(ctx, &myData)

		// Assert
		res := state.Retrieve(ctx)
		assert.Equal(t, myData.Val, res.Val)
	})

	t.Run("Should return nil if data is not found with custom key", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		state := state.NewState(
			state.WithCtxKey[MyData]("my-key"),
		)

		// Act
		res := state.Retrieve(ctx)

		// Assert
		assert.Empty(t, res)
	})
}
