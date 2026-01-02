package gemini_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/monarch-dev/monarch/internal/llm/gemini"
)

func TestGeminiClient_Init(t *testing.T) {
	c, err := gemini.NewClient("fake-key")
	assert.NoError(t, err)
	assert.NotNil(t, c)
}
