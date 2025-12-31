package project_test

import (
	"context"
	"testing"

	"github.com/monarch-dev/monarch/database"
	"github.com/monarch-dev/monarch/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Create(ctx context.Context, path string) (database.Project, error) {
	args := m.Called(ctx, path)
	if args.Get(0) == nil {
		return database.Project{}, args.Error(1)
	}
	return args.Get(0).(database.Project), args.Error(1)
}

func TestRegister_InvalidPath(t *testing.T) {
	svc := project.NewService(&MockStore{})
	_, err := svc.Register(context.Background(), "/invalid/path/123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path does not exist")
}
