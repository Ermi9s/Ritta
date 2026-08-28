package deploy

import (
	"fmt"
	"ritta/internal/config"
	"ritta/internal/env"
	"ritta/internal/logger"
	rittaSSH "ritta/internal/ssh"
	"time"
)

type Deployer struct {
	Config *config.Config
	SSH    *rittaSSH.Client
	log    *logger.Logger
}

func New(cfg *config.Config, sshClient *rittaSSH.Client, log *logger.Logger) *Deployer {
	return &Deployer{
		Config: cfg,
		SSH:    sshClient,
		log:    log,
	}
}

const actionDelay = 500 * time.Millisecond // just to make things epic

func (d *Deployer) Deploy(scanEnv bool) error {
	d.log.Success("Starting deployment...")
	time.Sleep(actionDelay)

	if err := d.runSetupScript(); err != nil {
		d.log.Errorf("Setup script failed: %v", err)
		return fmt.Errorf("setup script: %w", err)
	}
	time.Sleep(actionDelay)

	d.log.Success("Setup complete")

	if err := d.prepareSource(); err != nil {
		d.log.Errorf("Preparing source failed: %v", err)
		return fmt.Errorf("preparing source: %w", err)
	}
	time.Sleep(actionDelay)
	d.log.Success("Source ready")

	if err := env.Deploy(d.SSH, d.Config, scanEnv); err != nil {
		d.log.Errorf("Deploying environment failed: %v", err)
		return fmt.Errorf("deploying environment: %w", err)
	}
	time.Sleep(actionDelay)
	d.log.Success("Environment ready")

	if err := d.build(); err != nil {
		d.log.Errorf("Build failed: %v", err)
		return fmt.Errorf("building application: %w", err)
	}

	time.Sleep(actionDelay)
	d.log.Success("Build complete :)")

	if err := d.run(); err != nil {
		d.log.Errorf("Starting application failed: %v", err)
		return fmt.Errorf("starting application: %w", err)
	}

	time.Sleep(actionDelay)
	d.log.Success("Application started")

	if err := d.healthCheck(); err != nil {
		d.log.Errorf("Health check failed: %v", err)
		return fmt.Errorf("health check: %w", err)
	}

	if err := d.configureProxy(); err != nil {
		d.log.Errorf("Configuring reverse proxy failed: %v", err)
		return fmt.Errorf("configuring reverse proxy: %w", err)
	}
	time.Sleep(actionDelay)
	d.log.Success("Reverse proxy configured")


	if err := d.configureTLS(); err != nil {
		d.log.Errorf("Configuring TLS failed: %v", err)
		return fmt.Errorf("configuring tls: %w", err)
	}
	time.Sleep(actionDelay)
	d.log.Success("TLS configured")

	d.log.Success("Deployment complete! :)")

	return nil
}
