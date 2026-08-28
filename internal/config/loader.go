package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)


func LoadConfig(path string) (*Config, error) {
	targetFile := path
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		targetFile = filepath.Join(path, filename)
	}

	data, err := os.ReadFile(targetFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}