package project

import (
	"context"

	"github.com/monarch-dev/monarch/database"
)

type Store interface {
	Create(ctx context.Context, path string) (database.Project, error)
}
