package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type Client struct {
	mock.Mock
}

func (m *Client) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *Client) Close() error {
	args := m.Called()
	return args.Error(0)
}
