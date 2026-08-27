package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)


func LoadConfig(path string)(*Config, error) {
	file := filepath.Join(path, filename)
	data , err := os.ReadFile(file);
	if err != nil {
		return nil, err;
	}

	var cfg Config;
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err;
	}
	return &cfg, nil;
}