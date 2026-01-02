package tools_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/monarch-dev/monarch/database"
	"github.com/monarch-dev/monarch/mcp/tools"
	"github.com/stretchr/testify/assert"
)

type MockQuerier struct {
	database.Querier
	Task *database.Task
}

func (m *MockQuerier) GetTask(ctx context.Context, id pgtype.UUID) (database.Task, error) {
	if m.Task != nil {
		return *m.Task, nil
	}
	return database.Task{}, errors.New("not found")
}

func (m *MockQuerier) UpdateTaskStatus(ctx context.Context, arg database.UpdateTaskStatusParams) error {
	return nil
}

func (m *MockQuerier) IncrementTaskAttempt(ctx context.Context, id pgtype.UUID) (int32, error) {
	return 0, nil
}

func TestBuilder_Submit_CircuitBreaker(t *testing.T) {
	// Setup task with 5 attempts
	task := &database.Task{
		ID:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		AttemptCount: 5,
	}
	mockDB := &MockQuerier{Task: task}

	builder := tools.NewBuilder(mockDB, nil)

	args := tools.TaskArgs{TaskID: "00000000-0000-0000-0000-000000000001"}
	result, _, err := builder.SubmitAttemptHandler(ctx(), &mcp.CallToolRequest{}, args)

	assert.NoError(t, err) // Handler returns error in result, not as return value usually
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "Task Blocked")
}
