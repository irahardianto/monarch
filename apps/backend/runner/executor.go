package runner

import (
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
)

// ExecClient defines the interface for executing commands in containers.
// It matches a subset of the Docker client API.
type ExecClient interface {
	ContainerExecCreate(ctx context.Context, container string, config container.ExecOptions) (types.IDResponse, error)
	ContainerExecAttach(ctx context.Context, execID string, config container.ExecAttachOptions) (types.HijackedResponse, error)
	ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error)
}

type Executor struct {
	cli ExecClient
}

func NewExecutor(cli ExecClient) *Executor {
	return &Executor{cli: cli}
}

func (e *Executor) Run(ctx context.Context, containerID string, cmd []string) (string, string, int, error) {
	// 1. Create Exec
	cfg := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}
	resp, err := e.cli.ContainerExecCreate(ctx, containerID, cfg)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create exec: %w", err)
	}

	// 2. Attach
	attachResp, err := e.cli.ContainerExecAttach(ctx, resp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to attach exec: %w", err)
	}
	defer attachResp.Close()

	// 3. Capture Output
	var outBuf, errBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, attachResp.Reader)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to copy output: %w", err)
	}

	// 4. Inspect for Exit Code
	inspectResp, err := e.cli.ContainerExecInspect(ctx, resp.ID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to inspect exec: %w", err)
	}

	return outBuf.String(), errBuf.String(), inspectResp.ExitCode, nil
}
