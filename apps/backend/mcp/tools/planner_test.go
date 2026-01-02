package tools_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/monarch-dev/monarch/database"
	"github.com/monarch-dev/monarch/mcp/tools"
	"github.com/monarch-dev/monarch/project"
	"github.com/stretchr/testify/assert"
)

// MockStore for Project Service
type MockProjectStore struct{}

func (m *MockProjectStore) Create(ctx context.Context, path string) (database.Project, error) {
	return database.Project{}, nil
}

func (m *MockProjectStore) List(ctx context.Context) ([]database.Project, error) {
	return []database.Project{
		{Path: "/tmp/test", ID: pgtype.UUID{Valid: true}},
	}, nil
}

func TestPlanner_ListProjects(t *testing.T) {
	// Setup Mocks
	mockStore := &MockProjectStore{}
	svc := project.NewService(mockStore)
	planner := tools.NewPlanner(svc)

	// Call tool handler directly
	// We pass empty request as we don't need args
	result, _, err := planner.ListProjectsHandler(ctx(), &mcp.CallToolRequest{}, struct{}{})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// We expect a text result containing the project list
	assert.False(t, result.IsError)
}

func TestPlanner_SearchTasks(t *testing.T) {
	mockStore := &MockProjectStore{}
	svc := project.NewService(mockStore)
	planner := tools.NewPlanner(svc)

	args := struct {
		Query string `json:"query"`
	}{Query: "test"}

	result, _, err := planner.SearchTasksHandler(ctx(), &mcp.CallToolRequest{}, args)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "Task 1")
}

func ctx() context.Context {
	return context.Background()
}
