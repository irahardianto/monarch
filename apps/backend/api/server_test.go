package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/monarch-dev/monarch/api"
	"github.com/monarch-dev/monarch/config"
	"github.com/stretchr/testify/assert"
)

func TestServer_Health(t *testing.T) {
	cfg := &config.Config{Env: "test", Port: 8080}
	srv := api.NewServer(cfg, nil, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
