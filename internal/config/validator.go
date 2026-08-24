package config

import (
	"errors"
	"fmt"
)


func Validate(cfg *Config) error {
	if cfg.RemoteProjectRoot == "" {
		return errors.New("remote project root is required")
	}

	if cfg.LocalProjectRoot == "" {
		return errors.New("local project root is required")
	}

	if cfg.Server.Host == "" {
		return errors.New("server.host is required")
	}

	if cfg.Server.User == "" {
		return errors.New("server.user is required")
	}

	switch cfg.Source.Type {
	case "existing":
	case "git":
		if cfg.Source.Repository == "" {
			return errors.New(
				"source.repository is required for git source",
			)
		}
	default:
		return fmt.Errorf(
			"unsupported source type: %s",
			cfg.Source.Type,
		)
	}

	return nil
}