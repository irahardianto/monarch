package gates

type Config struct {
	Stack string `yaml:"stack"`
	Gates []Gate `yaml:"gates"`
}

type Gate struct {
	Name        string `yaml:"name"`
	Command     string `yaml:"command"` // For Standard gates
	Tier        string `yaml:"tier"`    // A, B, C
	Type        string `yaml:"type"`    // "standard" (default) or "llm_eval"
	Instruction string `yaml:"instruction"` // For LLM gates
	File        string `yaml:"file"`        // For LLM gates (target file)
}