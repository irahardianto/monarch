package parser_test

import (
	"testing"

	"github.com/monarch-dev/monarch/runner/parser"
	"github.com/stretchr/testify/assert"
)

func TestESLintParser_Parse(t *testing.T) {
	raw := []byte(`[{"filePath":"app.ts","messages":[{"ruleId":"no-console","severity":2,"message":"Unexpected console","line":10}]}]`)
	p := &parser.ESLintParser{}
	entries, err := p.Parse(raw)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "no-console", entries[0].RuleID)
	assert.Equal(t, parser.SeverityError, entries[0].Severity)
}
