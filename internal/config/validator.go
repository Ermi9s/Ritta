package config

import (
	"errors"
	"fmt"
	"strings"
)

func Validate(cfg *Config) error {
	if cfg == nil {
		return errors.New("configuration cannot be nil")
	}

	if cfg.RemoteProjectRoot == "" {
		return errors.New("remote project root is required")
	}

	if cfg.LocalProjectRoot == "" {
		return errors.New("local project root is required")
	}

	if cfg.Server.Host == "" {
		return errors.New("server host is required")
	}

	if cfg.Server.User == "" {
		return errors.New("server user is required")
	}

	if cfg.Server.Port != 0 && (cfg.Server.Port < 1 || cfg.Server.Port > 65535) {
		return fmt.Errorf("invalid server port: %d (must be between 1 and 65535)", cfg.Server.Port)
	}

	switch cfg.Source.Type {
	case "existing":
	case "git":
		if cfg.Source.Repository == "" {
			return errors.New("source.repository is required for git source")
		}
	default:
		return fmt.Errorf("unsupported source type: %s", cfg.Source.Type)
	}

	for i, domain := range cfg.Domains {
		if domain.Host == "" {
			return fmt.Errorf("domain at index %d has empty host", i)
		}
		if domain.Port < 1 || domain.Port > 65535 {
			return fmt.Errorf("domain %q has invalid port: %d", domain.Host, domain.Port)
		}
	}

	if cfg.Proxy != nil && cfg.Proxy.Provider != "" {
		provider := strings.ToLower(cfg.Proxy.Provider)
		if provider != "nginx" {
			return fmt.Errorf("unsupported proxy provider: %s (supported: nginx)", cfg.Proxy.Provider)
		}
	}

	if cfg.TLS != nil && cfg.TLS.Provider != "" {
		provider := strings.ToLower(cfg.TLS.Provider)
		if provider != "letsencrypt" && provider != "certbot" {
			return fmt.Errorf("unsupported TLS provider: %s (supported: letsencrypt, certbot)", cfg.TLS.Provider)
		}
	}

	return nil
}