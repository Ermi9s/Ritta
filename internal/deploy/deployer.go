package deploy

import (
	"fmt"
	"ritta/internal/config"
	"ritta/internal/env"
	rittaSSH "ritta/internal/ssh"
)

type Deployer struct {
	Config *config.Config
	SSH    *rittaSSH.Client
}

func New(cfg *config.Config, sshClient *rittaSSH.Client) *Deployer {
	return &Deployer{
		Config: cfg,
		SSH:    sshClient,
	}
}

func (d *Deployer) Deploy() error {
	fmt.Println("Preparing deployment...")

	if err := d.runSetupScript(); err != nil {
		return fmt.Errorf("setup script: %w", err)
	}

	fmt.Println(":) Setup complete")

	if err := d.prepareSource(); err != nil {
		return fmt.Errorf("preparing source: %w", err)
	}

	fmt.Println(":) Source ready")

	if err := env.Deploy(d.SSH, d.Config); err != nil {
		return fmt.Errorf("deploying environment: %w", err)
	}

	fmt.Println(":) Environment ready")

	if err := d.build(); err != nil {
		return fmt.Errorf("building application: %w", err)
	}

	fmt.Println(":) Build complete")

	if err := d.run(); err != nil {
		return fmt.Errorf("starting application: %w", err)
	}

	fmt.Println(":) Application started")

	if err := d.healthCheck(); err != nil {
		return fmt.Errorf("health check: %w", err)
	}

	if err := d.configureProxy(); err != nil {
		return fmt.Errorf("configuring reverse proxy: %w", err)
	}

	if err := d.configureTLS(); err != nil {
		return fmt.Errorf("configuring tls: %w", err)
	}

	fmt.Println()
	fmt.Println(":) Deployment complete")

	return nil
}

