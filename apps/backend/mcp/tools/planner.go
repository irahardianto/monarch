package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/monarch-dev/monarch/project"
)

type Planner struct {
	projectService *project.Service
}

func NewPlanner(svc *project.Service) *Planner {
	return &Planner{projectService: svc}
}

func (p *Planner) Register(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_projects",
		Description: "List all registered projects",
	}, p.ListProjectsHandler)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_past_tasks",
		Description: "Search past tasks using vector embeddings",
	}, p.SearchTasksHandler)
}

func (p *Planner) SearchTasksHandler(ctx context.Context, req *mcp.CallToolRequest, args struct {
	Query string `json:"query"`
}) (*mcp.CallToolResult, any, error) {
	// Mock implementation for now
	results := []string{}
	if args.Query == "test" {
		results = []string{"Task 1: Test Task", "Task 2: Another Test"}
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("failed to marshal results: %v", err),
				},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(data),
			},
		},
	}, nil, nil
}


func (p *Planner) ListProjectsHandler(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	projects, err := p.projectService.List(ctx)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: err.Error(),
				},
			},
		}, nil, nil
	}

	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("failed to marshal projects: %v", err),
				},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(data),
			},
		},
	}, nil, nil
}