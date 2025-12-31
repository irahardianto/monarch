package runner_test

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/monarch-dev/monarch/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReapZombies_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)
	defer cli.Close()

	// 1. Create a dummy container with label
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "hello"},
		Labels: map[string]string{
			"monarch.managed": "true",
		},
	}, nil, nil, nil, "monarch-test-zombie")
	
	// Handle case where container might already exist from failed previous run
	if err != nil {
		// Try to remove it and recreate if it exists? 
		// For now assume unique name or just proceed if clean env
		// Better: cleanup before test
		_ = cli.ContainerRemove(ctx, "monarch-test-zombie", container.RemoveOptions{Force: true})
		resp, err = cli.ContainerCreate(ctx, &container.Config{
			Image: "alpine",
			Cmd:   []string{"echo", "hello"},
			Labels: map[string]string{
				"monarch.managed": "true",
			},
		}, nil, nil, nil, "monarch-test-zombie")
		require.NoError(t, err)
	}

	// 2. Reap
	count, err := runner.ReapZombies(ctx, cli)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// 3. Verify gone
	_, err = cli.ContainerInspect(ctx, resp.ID)
	assert.Error(t, err)
	assert.True(t, client.IsErrNotFound(err))
}
