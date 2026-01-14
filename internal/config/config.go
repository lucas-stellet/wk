// Package config handles parsing of .wk.yaml configuration files.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigFileName is the default configuration file name.
const ConfigFileName = ".wk.yaml"

// Config represents the wk configuration for a project.
type Config struct {
	// Copy lists files and directories to copy from source to new worktree.
	Copy []string `yaml:"copy"`
	// PostHooks lists commands to run after creating the worktree.
	PostHooks []string `yaml:"post_hooks"`
}

// Load reads and parses a configuration file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// FindConfig searches for .wk.yaml starting from dir and walking up to root.
func FindConfig(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	for {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}
