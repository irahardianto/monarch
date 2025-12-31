package runner

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// DockerClient defines the subset of client.Client methods we use
// This interface allows for mocking in tests
type DockerClient interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
}

func ReapZombies(ctx context.Context, cli DockerClient) (int, error) {
	args := filters.NewArgs()
	args.Add("label", "monarch.managed=true")

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return 0, err
	}

	count := 0
	for _, c := range containers {
		if err := cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err == nil {
			count++
		}
	}
	return count, nil
}
