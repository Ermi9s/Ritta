package proxy

import (
	"ritta/internal/config"
	proxyproviders "ritta/internal/proxy/proxy.providers"
	"ritta/internal/ssh"
)


func NewReverseProxy(client *ssh.Client, cfg *config.Config) ProxyInterface {
	if cfg.Proxy == nil {
		return nil
	}
	switch cfg.Proxy.Provider {
	case "Nginx":
		return proxyproviders.NewNginx(client);
	}


	return nil;
}


