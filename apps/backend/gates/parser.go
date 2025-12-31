package gates

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFile = ".monarch/gates.yaml"

func DetectStack(root string) (*Config, error) {
	// 1. Check for explicit config
	configPath := filepath.Join(root, configFile)
	if _, err := os.Stat(configPath); err == nil {
		return parseFile(configPath)
	}

	// 2. Auto-detect
	stack := "unknown"

	if exists(filepath.Join(root, "go.mod")) {
		stack = "go"
	} else if exists(filepath.Join(root, "package.json")) {
		stack = "node"
	} else if exists(filepath.Join(root, "requirements.txt")) || exists(filepath.Join(root, "pyproject.toml")) {
		stack = "python"
	}

	return &Config{Stack: stack}, nil
}

func parseFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
