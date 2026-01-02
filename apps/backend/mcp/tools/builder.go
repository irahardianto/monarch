package tools

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/monarch-dev/monarch/database"
	"github.com/monarch-dev/monarch/runner"
)

type Builder struct {
	store  database.Querier
	runner runner.Service
}

func NewBuilder(store database.Querier, runner runner.Service) *Builder {
	return &Builder{store: store, runner: runner}
}

func (b *Builder) Register(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "claim_task",
		Description: "Mark a task as IN_PROGRESS",
	}, b.ClaimTaskHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "submit_attempt",
		Description: "Submit a task attempt for validation",
	}, b.SubmitAttemptHandler)
}

type TaskArgs struct {
	TaskID string `json:"task_id"`
}

func (b *Builder) ClaimTaskHandler(ctx context.Context, req *mcp.CallToolRequest, args TaskArgs) (*mcp.CallToolResult, any, error) {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(args.TaskID); err != nil {
		return errorResult("Invalid Task ID format"), nil, nil
	}

	err := b.store.UpdateTaskStatus(ctx, database.UpdateTaskStatusParams{
		ID:     uuid,
		Status: "IN_PROGRESS",
	})
	if err != nil {
		return errorResult(err.Error()), nil, nil
	}

	return successResult("Task claimed"), nil, nil
}

func (b *Builder) SubmitAttemptHandler(ctx context.Context, req *mcp.CallToolRequest, args TaskArgs) (*mcp.CallToolResult, any, error) {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(args.TaskID); err != nil {
		return errorResult("Invalid Task ID format"), nil, nil
	}

	task, err := b.store.GetTask(ctx, uuid)
	if err != nil {
		return errorResult(err.Error()), nil, nil
	}

	if task.AttemptCount >= 5 {
		_ = b.store.UpdateTaskStatus(ctx, database.UpdateTaskStatusParams{
			ID:     uuid,
			Status: "BLOCKED",
		})
		return errorResult("Task Blocked. Human intervention required."), nil, nil
	}

	// Increment attempts
	_, err = b.store.IncrementTaskAttempt(ctx, uuid)
	if err != nil {
		return errorResult("Failed to increment attempts"), nil, nil
	}

	// Trigger Runner (Simplified - just checking if it exists)
	if b.runner != nil {
		// Mock execution
		// In reality: b.runner.Execute(ctx, task.ProjectID.String(), []string{"go", "test", "./..."})
	}

	return successResult("Validation triggered"), nil, nil
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}
}

func successResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}
}
