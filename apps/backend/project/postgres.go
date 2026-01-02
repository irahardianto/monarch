package project

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/monarch-dev/monarch/database"
)

type PostgresStore struct {
	q *database.Queries
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{q: database.New(db)}
}

func (s *PostgresStore) Create(ctx context.Context, path string) (database.Project, error) {
	return s.q.CreateProject(ctx, path)
}

func (s *PostgresStore) List(ctx context.Context) ([]database.Project, error) {
	return s.q.ListProjects(ctx)
}
