package gates

type Config struct {
	Stack string `yaml:"stack"`
	Gates []Gate `yaml:"gates"`
}

type Gate struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Tier    string `yaml:"tier"` // A, B, C
}
