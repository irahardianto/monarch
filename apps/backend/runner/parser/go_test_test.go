package parser_test

import (
	"testing"

	"github.com/monarch-dev/monarch/runner/parser"
	"github.com/stretchr/testify/assert"
)

func TestGoTestParser_Parse(t *testing.T) {
	raw := []byte(`{"Action":"fail","Package":"m/test","Test":"TestFoo","Output":"FAIL: TestFoo (0.00s)\n"}`)
	p := &parser.GoTestParser{}
	entries, err := p.Parse(raw)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "FAIL: TestFoo (0.00s)", entries[0].Message)
	assert.Equal(t, "m/test", entries[0].File)
	assert.Equal(t, parser.SeverityError, entries[0].Severity)
}
