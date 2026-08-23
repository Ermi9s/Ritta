package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

func LoadConfig(file string)(*Config, error) {
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