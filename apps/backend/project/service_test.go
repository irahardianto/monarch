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
	proj database.Project
	err  error
}

func (m *MockStore) Create(ctx context.Context, path string) (database.Project, error) {
	if m.err != nil {
		return database.Project{}, m.err
	}
	return m.proj, nil
}

func (m *MockStore) List(ctx context.Context) ([]database.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []database.Project{m.proj}, nil
}

func TestRegister_InvalidPath(t *testing.T) {
	svc := project.NewService(&MockStore{})
	_, err := svc.Register(context.Background(), "/invalid/path/123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path does not exist")
}
