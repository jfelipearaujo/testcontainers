package state

import "context"

// CtxKeyType is a type that can be used as a key for a context.Context
type CtxKeyType string

// State is a type that can be used to store data in a context.Context
// It is useful for storing data that needs to be shared between tests
type State[T any] struct {
	CtxKey CtxKeyType
}

// StateOption is a type that can be used to configure a State
type StateOption[T any] func(*State[T])

// WithCtxKey is a StateOption that sets the key for the context.Context
func WithCtxKey[T any](ctxKey CtxKeyType) StateOption[T] {
	return func(state *State[T]) {
		state.CtxKey = ctxKey
	}
}

// NewState creates a new State
func NewState[T any](opts ...StateOption[T]) *State[T] {
	state := &State[T]{
		CtxKey: "default",
	}

	for _, opt := range opts {
		opt(state)
	}

	return state
}

// Enrich enriches the context with the data
func (state *State[T]) Enrich(ctx context.Context, data *T) context.Context {
	return context.WithValue(ctx, state.CtxKey, data)
}

// Retrieve retrieves the data from the context
func (state *State[T]) Retrieve(ctx context.Context) *T {
	data, ok := ctx.Value(state.CtxKey).(*T)
	if !ok {
		return new(T)
	}
	return data
}
