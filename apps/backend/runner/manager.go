package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// ExtendedDockerClient adds ContainerCreate and ContainerStart to our interface
type ExtendedDockerClient interface {
	DockerClient
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
}

type Manager struct {
	cli ExtendedDockerClient
	mu  sync.RWMutex
	// runners maps ProjectID -> Stack -> ContainerID
	runners map[string]map[string]string
	// lastUsed maps ContainerID -> timestamp
	lastUsed map[string]time.Time
}

func NewManager(cli ExtendedDockerClient) *Manager {
	return &Manager{
		cli:      cli,
		runners:  make(map[string]map[string]string),
		lastUsed: make(map[string]time.Time),
	}
}

// StartMonitor runs a background loop to clean up idle containers
func (m *Manager) StartMonitor(ctx context.Context, interval, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.checkIdle(ctx, timeout)
			}
		}
	}()
}

func (m *Manager) checkIdle(ctx context.Context, timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for pid, stacks := range m.runners {
		for stack, cid := range stacks {
			if last, ok := m.lastUsed[cid]; ok {
				if now.Sub(last) > timeout {
					// Stop container
					// We ignore error here as it's a background cleanup
					// and we want to proceed with cleanup from map
					_ = m.cli.ContainerStop(ctx, cid, container.StopOptions{})
					
					// Remove from maps
					delete(stacks, stack)
					delete(m.lastUsed, cid)
				}
			}
		}
		// Clean up empty project maps
		if len(stacks) == 0 {
			delete(m.runners, pid)
		}
	}
}

func (m *Manager) GetOrStart(ctx context.Context, projectID, stack string) (string, error) {
	m.mu.RLock()
	if stacks, ok := m.runners[projectID]; ok {
		if id, ok := stacks[stack]; ok {
			m.mu.RUnlock()
			m.touch(id)
			return id, nil
		}
	}
	m.mu.RUnlock()

	return m.startContainer(ctx, projectID, stack)
}

func (m *Manager) touch(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastUsed[id] = time.Now()
}

func (m *Manager) startContainer(ctx context.Context, projectID, stack string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check locking
	if stacks, ok := m.runners[projectID]; ok {
		if id, ok := stacks[stack]; ok {
			m.lastUsed[id] = time.Now()
			return id, nil
		}
	}

	// For MVP, we use a placeholder image. Real implementation would map stack -> image
	image := "alpine" 
	cmd := []string{"sleep", "infinity"}

	resp, err := m.cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   cmd,
		Labels: map[string]string{
			"monarch.managed": "true",
			"monarch.project": projectID,
			"monarch.stack":   stack,
		},
	}, nil, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	if err := m.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	if _, ok := m.runners[projectID]; !ok {
		m.runners[projectID] = make(map[string]string)
	}
	m.runners[projectID][stack] = resp.ID
	m.lastUsed[resp.ID] = time.Now()

	return resp.ID, nil
}
