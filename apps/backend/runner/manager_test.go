package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/monarch-dev/monarch/runner"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDockerClient implements runner.ExtendedDockerClient
type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	args := m.Called(ctx, options)
	return args.Get(0).([]types.Container), args.Error(1)
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func TestGetOrStart_StartsNewContainer(t *testing.T) {
	mockCli := new(MockDockerClient)
	mgr := runner.NewManager(mockCli)
	ctx := context.Background()

	// Expectation: ContainerCreate called with correct labels
	mockCli.On("ContainerCreate", ctx, mock.MatchedBy(func(cfg *container.Config) bool {
		return cfg.Labels["monarch.managed"] == "true" &&
			cfg.Labels["monarch.project"] == "proj-1" &&
			cfg.Labels["monarch.stack"] == "python-3.11"
	}), (*container.HostConfig)(nil), (*network.NetworkingConfig)(nil), (*v1.Platform)(nil), "").
		Return(container.CreateResponse{ID: "new-container-123"}, nil)

	mockCli.On("ContainerStart", ctx, "new-container-123", container.StartOptions{}).
		Return(nil)

	id, err := mgr.GetOrStart(ctx, "proj-1", "python-3.11")
	assert.NoError(t, err)
	assert.Equal(t, "new-container-123", id)

	mockCli.AssertExpectations(t)
}

func TestGetOrStart_ReturnsExistingContainer(t *testing.T) {
	mockCli := new(MockDockerClient)
	mgr := runner.NewManager(mockCli)
	ctx := context.Background()

	// Pre-seed map logic (simulated by calling GetOrStart once)
	// First call setup
	mockCli.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "existing-id"}, nil).Once()
	mockCli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()

	// Priming call
	_, _ = mgr.GetOrStart(ctx, "proj-1", "stack-1")

	// Second call - should NOT call Docker API
	id, err := mgr.GetOrStart(ctx, "proj-1", "stack-1")
	assert.NoError(t, err)
	assert.Equal(t, "existing-id", id)

	mockCli.AssertNumberOfCalls(t, "ContainerCreate", 1)
}

func TestMonitor_StopsIdleContainers(t *testing.T) {
	mockCli := new(MockDockerClient)
	mgr := runner.NewManager(mockCli)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Start a container
	mockCli.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "idle-container"}, nil).Once()
	mockCli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()

	_, err := mgr.GetOrStart(ctx, "proj-1", "stack-1")
	require.NoError(t, err)

	// 2. Expect Stop
	mockCli.On("ContainerStop", mock.Anything, "idle-container", mock.Anything).
		Return(nil).Once()

	// 3. Start Monitor with short interval
	mgr.StartMonitor(ctx, 10*time.Millisecond, 50*time.Millisecond)

	// 4. Wait for idle (simulated by sleep > timeout)
	time.Sleep(100 * time.Millisecond)

	// 5. Verify it's gone from map (internal state check via public behavior)
	// If we call GetOrStart again, it should try to Create again, not return existing
	mockCli.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(container.CreateResponse{ID: "new-container"}, nil).Once()
	mockCli.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()

	id, err := mgr.GetOrStart(ctx, "proj-1", "stack-1")
	assert.NoError(t, err)
	assert.Equal(t, "new-container", id)

	mockCli.AssertExpectations(t)
}
