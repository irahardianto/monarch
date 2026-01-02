package parser

var ruleHints = map[string]string{
	"no-console": "Console logs are forbidden in production. Use a structured logger.",
	"G101":       "Potential hardcoded credential. Use environment variables.",
	// Add more as needed
}

func Enrich(entry *LogEntry) {
	if hint, ok := ruleHints[entry.RuleID]; ok {
		entry.Hint = hint
	}
}
