package config_test

import (
	"os"
	"testing"

	"github.com/monarch-dev/monarch/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "development", cfg.Env)
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("MONARCH_PORT", "8080")
	os.Setenv("MONARCH_ENV", "production")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	defer os.Clearenv()

	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.DB)
}
