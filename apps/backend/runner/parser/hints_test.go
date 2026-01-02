package parser_test

import (
	"testing"

	"github.com/monarch-dev/monarch/runner/parser"
	"github.com/stretchr/testify/assert"
)

func TestEnrich(t *testing.T) {
	entry := &parser.LogEntry{RuleID: "no-console"}
	parser.Enrich(entry)
	assert.NotEmpty(t, entry.Hint)
	assert.Contains(t, entry.Hint, "Console")
}
