package project

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/monarch-dev/monarch/database"
	"github.com/monarch-dev/monarch/gates"
)

type Project struct {
	*database.Project
	HasGit bool          `json:"has_git"`
	Config *gates.Config `json:"config"`
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Register(ctx context.Context, path string) (*Project, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("path does not exist")
	}

	// Check Git
	hasGit := false
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		hasGit = true
	}

	// Detect Gates
	config, err := gates.DetectStack(path)
	if err != nil {
		return nil, err
	}

	// Save to DB
	dbProj, err := s.store.Create(ctx, path)
	if err != nil {
		return nil, err
	}

	return &Project{
		Project: &dbProj,
		HasGit:  hasGit,
		Config:  config,
	}, nil
}
