package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/monarch-dev/monarch/gates"
	"github.com/monarch-dev/monarch/runner/eval"
)

type Service interface {
	Execute(ctx context.Context, projectID string, cmd []string) (string, error)
	RunGate(ctx context.Context, projectID string, gate gates.Gate) error
}

type RunnerService struct {
	manager    *Manager
	executor   *Executor
	evalEngine *eval.Engine
}

func NewService(manager *Manager, executor *Executor, evalEngine *eval.Engine) *RunnerService {
	return &RunnerService{
		manager:    manager,
		executor:   executor,
		evalEngine: evalEngine,
	}
}

func (s *RunnerService) Execute(ctx context.Context, projectID string, cmd []string) (string, error) {
	// Retrieve container (assuming 'default' stack for raw execute, or pass stack in)
	containerID, err := s.manager.GetOrStart(ctx, projectID, "default")
	if err != nil {
		return "", err
	}

	stdout, stderr, exitCode, err := s.executor.Run(ctx, containerID, cmd)
	if err != nil {
		return "", err
	}

	if exitCode != 0 {
		return "", fmt.Errorf("execution failed (exit %d): %s", exitCode, stderr)
	}

	return stdout, nil
}

func (s *RunnerService) RunGate(ctx context.Context, projectID string, gate gates.Gate) error {
	if gate.Type == "llm_eval" {
		// Default to Snapshot mode for now as per plan focus
		res, err := s.evalEngine.EvaluateSnapshot(ctx, gate.File, gate.Instruction)
		if err != nil {
			return err
		}
		
		// Basic validation of result
		if strings.Contains(strings.ToUpper(res), "FAIL") {
			return fmt.Errorf("LLM evaluation failed: %s", res)
		}
		return nil
	}

	// Standard Execution
	// Determine stack from gate or default
	// Assuming manager needs stack. Gate config has stack at top level, passed down?
	// For now using "default" or "go" as placeholder if not in Gate struct.
	// We added Tier, Command, Name. Not Stack in Gate. Stack is in Config.
	stack := "default" 
	
	containerID, err := s.manager.GetOrStart(ctx, projectID, stack)
	if err != nil {
		return err
	}

	cmd := strings.Fields(gate.Command)
	_, stderr, exitCode, err := s.executor.Run(ctx, containerID, cmd)
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return fmt.Errorf("gate %s failed: %s", gate.Name, stderr)
	}

	return nil
}