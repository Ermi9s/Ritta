package proxy

import (
	"ritta/internal/config"
	rittaSSH "ritta/internal/ssh"
)


type ProxyInterface interface {
	EnsureInstalled() error
	ConfigureDomain(domain config.Domain) error 
	Configure(domains []config.Domain) error 
	Reload() error 
	Test() error 
}

type TLSProvider interface {
	Configure(client *rittaSSH.Client, cfg *config.Config) error
}