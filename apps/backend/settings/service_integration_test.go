package settings_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/monarch-dev/monarch/settings"
	"github.com/monarch-dev/monarch/database"
)

func setupTestDB(t *testing.T) database.DBTX {
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		t.Skip("TEST_DB_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		pool.Close()
	})
	return pool
}

func TestSettingsService_Integration(t *testing.T) {
	db := setupTestDB(t) 
	
	// Key for testing (32 bytes)
	encKey := []byte("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	
	svc := settings.NewService(db, encKey)

	ctx := context.Background()
	
	// Test Save Encrypted
	err := svc.Set(ctx, "GEMINI_API_KEY", "secret-value", true)
	require.NoError(t, err)

	// Test Get
	val, err := svc.Get(ctx, "GEMINI_API_KEY")
	require.NoError(t, err)
	require.Equal(t, "secret-value", val)
}
