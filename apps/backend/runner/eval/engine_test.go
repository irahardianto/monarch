package eval_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/monarch-dev/monarch/runner/eval"
	"github.com/monarch-dev/monarch/internal/llm/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestEvaluateSnapshot_Limit(t *testing.T) {
	mockLLM := new(mocks.Client)
	engine := eval.NewEngine(mockLLM, 100) // 100 bytes limit

	// Create dummy large file in temp dir
	tmpDir := t.TempDir()
	largeFile := filepath.Join(tmpDir, "large.txt")
	err := os.WriteFile(largeFile, make([]byte, 101), 0644)
	require.NoError(t, err)

	_, err = engine.EvaluateSnapshot(context.Background(), largeFile, "Check this")
	require.ErrorContains(t, err, "file size 101 exceeds limit 100")
}

func TestEvaluateSnapshot_Success(t *testing.T) {
	mockLLM := new(mocks.Client)
	engine := eval.NewEngine(mockLLM, 1000)

	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "small.txt")
	content := []byte("hello world")
	err := os.WriteFile(file, content, 0644)
	require.NoError(t, err)

	// Expectation
	mockLLM.On("Generate", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		return len(prompt) > len("hello world")
	})).Return("looks good", nil)

	res, err := engine.EvaluateSnapshot(context.Background(), file, "instruction")
	require.NoError(t, err)
	require.Equal(t, "looks good", res)
}
