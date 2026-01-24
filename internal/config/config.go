package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Prefix     string `toml:"prefix"`
	IDLength   int    `toml:"id_length"`
	projectDir string
	pebblesDir string
}

const DefaultPrefix = "peb"
const DefaultIDLength = 4

var ErrNoPebblesDir = errors.New("no .pebbles directory found (did you run 'peb init'?)")

func Load() (*Config, error) {
	dir, err := findPebblesDir()
	if err != nil {
		return nil, err
	}
	cfg, err := loadConfig(dir)
	if err != nil {
		return nil, err
	}
	cfg.projectDir = filepath.Dir(dir)
	cfg.pebblesDir = dir
	return cfg, nil
}

func findPebblesDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		pebblesDir := filepath.Join(dir, ".pebbles")
		if info, err := os.Stat(pebblesDir); err == nil && info.IsDir() {
			return pebblesDir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoPebblesDir
		}
		dir = parent
	}
}

func loadConfig(pebblesDir string) (*Config, error) {
	configPath := filepath.Join(pebblesDir, "config.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	cfg := &Config{
		Prefix:   DefaultPrefix,
		IDLength: DefaultIDLength,
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

func (c *Config) PebblesDir() string {
	return c.pebblesDir
}

func DefaultConfigContent() string {
	return fmt.Sprintf(`# Pebbles configuration
prefix = "%s"
id_length = %d
`, DefaultPrefix, DefaultIDLength)
}
