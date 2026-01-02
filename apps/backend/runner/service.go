package runner

import (
	"context"
)

type Service interface {
	Execute(ctx context.Context, projectID string, cmd []string) (string, error)
}
