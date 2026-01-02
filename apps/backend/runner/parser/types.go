package parser

import "errors"

// ErrSystemFailure indicates the tool output was malformed or crashed.
// This triggers the "Fail Closed" mechanism.
var ErrSystemFailure = errors.New("system failure: tool output malformed")

type Severity string

const (
	SeverityInfo    Severity = "INFO"
	SeverityWarning Severity = "WARNING"
	SeverityError   Severity = "ERROR"
)

type LogEntry struct {
	Severity Severity `json:"severity"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Message  string   `json:"message"`
	Tool     string   `json:"tool"`
	RuleID   string   `json:"rule_id,omitempty"` // e.g., "G101"
	Hint     string   `json:"hint,omitempty"`    // Enriched advice
}

type Parser interface {
	// Parse converts raw tool output into structured entries.
	// Must return ErrSystemFailure if output is unparseable.
	Parse(raw []byte) ([]LogEntry, error)
}
