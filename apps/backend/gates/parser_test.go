package gates_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monarch-dev/monarch/gates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectStack_AutoDetect(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  string
	}{
		{"Go", []string{"go.mod"}, "go"},
		{"Node", []string{"package.json"}, "node"},
		{"PythonReq", []string{"requirements.txt"}, "python"},
		{"PythonToml", []string{"pyproject.toml"}, "python"},
		{"Unknown", []string{}, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			for _, f := range tt.files {
				err := os.WriteFile(filepath.Join(tmp, f), []byte{}, 0644)
				require.NoError(t, err)
			}

			cfg, err := gates.DetectStack(tmp)
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.Stack)
		})
	}
}

func TestDetectStack_ExplicitConfig(t *testing.T) {
	tmp := t.TempDir()
	monarchDir := filepath.Join(tmp, ".monarch")
	require.NoError(t, os.Mkdir(monarchDir, 0755))

	yamlContent := `
stack: rust
gates:
  - name: build
    command: cargo build
    tier: A
`
	err := os.WriteFile(filepath.Join(monarchDir, "gates.yaml"), []byte(yamlContent), 0644)
	require.NoError(t, err)

	cfg, err := gates.DetectStack(tmp)
	require.NoError(t, err)
	assert.Equal(t, "rust", cfg.Stack)
	assert.Len(t, cfg.Gates, 1)
	assert.Equal(t, "build", cfg.Gates[0].Name)
}
