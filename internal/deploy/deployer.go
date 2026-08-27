package deploy

import (
	"fmt"
	"ritta/internal/config"
	"ritta/internal/env"
	rittaSSH "ritta/internal/ssh"
	"ritta/internal/ui"
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
	ui.Success("Starting deployment...\n")

	if err := d.runSetupScript(); err != nil {
		return fmt.Errorf("setup script: %w", err)
	}

	ui.Success("Setup complete")

	if err := d.prepareSource(); err != nil {
		return fmt.Errorf("preparing source: %w", err)
	}

	ui.Success("Source ready")

	if err := env.Deploy(d.SSH, d.Config); err != nil {
		return fmt.Errorf("deploying environment: %w", err)
	}

	ui.Success("Environment ready")

	if err := d.build(); err != nil {
		return fmt.Errorf("building application: %w", err)
	}

	ui.Success("Build complete :)")

	if err := d.run(); err != nil {
		return fmt.Errorf("starting application: %w", err)
	}

	ui.Success("Application started")

	if err := d.healthCheck(); err != nil {
		return fmt.Errorf("health check: %w", err)
	}

	if err := d.configureProxy(); err != nil {
		return fmt.Errorf("configuring reverse proxy: %w", err)
	}

	if err := d.configureTLS(); err != nil {
		return fmt.Errorf("configuring tls: %w", err)
	}

	ui.Success("\n\nDeployment complete")

	return nil
}

