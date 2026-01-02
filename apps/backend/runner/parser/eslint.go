package parser

import "encoding/json"

type ESLintParser struct{}

type eslintFile struct {
	FilePath string          `json:"filePath"`
	Messages []eslintMessage `json:"messages"`
}

type eslintMessage struct {
	RuleID   string `json:"ruleId"`
	Severity int    `json:"severity"` // 1=Warning, 2=Error
	Message  string `json:"message"`
	Line     int    `json:"line"`
}

func (p *ESLintParser) Parse(raw []byte) ([]LogEntry, error) {
	var files []eslintFile
	if err := json.Unmarshal(raw, &files); err != nil {
		return nil, ErrSystemFailure
	}

	var entries []LogEntry
	for _, file := range files {
		for _, msg := range file.Messages {
			severity := SeverityWarning
			if msg.Severity == 2 {
				severity = SeverityError
			}

			entries = append(entries, LogEntry{
				Severity: severity,
				File:     file.FilePath,
				Line:     msg.Line,
				Message:  msg.Message,
				Tool:     "eslint",
				RuleID:   msg.RuleID,
			})
		}
	}

	return entries, nil
}
