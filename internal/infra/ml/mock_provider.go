package ml

import (
	"context"
	"time"
)

// MockProvider is a simple provider for testing and building the UI flow
type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) CompareFaces(ctx context.Context, face1, face2 []byte) (float64, error) {
	// Simulate some processing time
	select {
	case <-time.After(800 * time.Millisecond):
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	// For mock, just return a random high score to allow testing the "Success" flow
	// or return 0.95 consistently for a "stable" mock
	return 0.95, nil
}

func (p *MockProvider) Close() error {
	return nil
}
