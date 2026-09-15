package proxy

import (
	"strings"

	"ritta/internal/config"
	"ritta/internal/logger"
	proxyproviders "ritta/internal/proxy/proxy.providers"
	"ritta/internal/ssh"
)

func NewReverseProxy(client *ssh.Client, cfg *config.Config, log *logger.Logger) ProxyInterface {
	if cfg.Proxy == nil {
		return nil
	}
	switch strings.ToLower(cfg.Proxy.Provider) {
	case "nginx":
		return proxyproviders.NewNginx(client, log)
	}

	return nil
}
