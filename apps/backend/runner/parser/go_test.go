package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
)

type GoTestParser struct{}

type goTestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

func (p *GoTestParser) Parse(raw []byte) ([]LogEntry, error) {
	var entries []LogEntry
	scanner := bufio.NewScanner(bytes.NewReader(raw))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var event goTestEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, ErrSystemFailure
		}

		if event.Action == "fail" && event.Test != "" {
			entries = append(entries, LogEntry{
				Severity: SeverityError,
				Message:  strings.TrimSpace(event.Output),
				Tool:     "go test",
				File:     event.Package, // Best effort context
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, ErrSystemFailure
	}

	return entries, nil
}
