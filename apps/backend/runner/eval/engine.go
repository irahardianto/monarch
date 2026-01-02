package eval

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"github.com/monarch-dev/monarch/internal/llm"
)

type Engine struct {
	llm       llm.Client
	sizeLimit int64
}

func NewEngine(llm llm.Client, sizeLimit int64) *Engine {
	return &Engine{llm: llm, sizeLimit: sizeLimit}
}

func (e *Engine) EvaluateSnapshot(ctx context.Context, path string, instruction string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > e.sizeLimit {
		return "", fmt.Errorf("file size %d exceeds limit %d", info.Size(), e.sizeLimit)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf("<file path=\"%s\">\n%s\n</file>\n\n<instruction>\n%s\n</instruction>", path, string(content), instruction)
	
	return e.llm.Generate(ctx, prompt)
}

func (e *Engine) EvaluateDiff(ctx context.Context, path string, instruction string) (string, error) {
	// Ensure path exists or handle git error
	cmd := exec.CommandContext(ctx, "git", "diff", path)
	diff, err := cmd.Output()
	if err != nil {
		// If git fails, maybe try to check if it's untracked?
		// For now return error
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	
	if len(diff) == 0 {
		return "No changes detected", nil
	}

	prompt := fmt.Sprintf("<diff path=\"%s\">\n%s\n</diff>\n\n<instruction>\n%s\n</instruction>", path, string(diff), instruction)
	
	return e.llm.Generate(ctx, prompt)
}
