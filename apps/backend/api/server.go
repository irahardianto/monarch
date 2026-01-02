package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/monarch-dev/monarch/config"
	"github.com/monarch-dev/monarch/project"
)

type Server struct {
	mux     *http.ServeMux
	handler http.Handler
	db      *pgxpool.Pool
	cfg     *config.Config
	projSvc *project.Service
	sse     *mcp.SSEHandler
}

func NewServer(cfg *config.Config, db *pgxpool.Pool, projSvc *project.Service, sse *mcp.SSEHandler) *Server {
	s := &Server{
		mux:     http.NewServeMux(),
		db:      db,
		cfg:     cfg,
		projSvc: projSvc,
		sse:     sse,
	}
	s.routes()
	s.handler = s.recoverer(s.logger(s.mux))
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
