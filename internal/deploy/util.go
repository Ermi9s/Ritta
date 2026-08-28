package deploy

import (
	"fmt"
	"ritta/internal/proxy"
	"os"
	"path/filepath"
)


func (d *Deployer) configureProxy() error {
	if len(d.Config.Domains) == 0 {
		d.log.Info("No domains configured")
		return nil
	}

	proxyProvider := proxy.NewReverseProxy(d.SSH, d.Config)

	if err := proxyProvider.Configure(d.Config.Domains); err != nil {
		return fmt.Errorf("configuring proxy %w", err);
	}

	return nil
}

func (d *Deployer) configureTLS() error {
	if d.Config.TLS.Provider == "" {
		d.log.Info("TLS not configured!")
		return nil
	}
	tlsprovider, err := proxy.NewTLSProvider(d.SSH, d.Config);
	if err != nil {
		return fmt.Errorf("creating tls provider: %w", err);
	}
	
	if err := tlsprovider.Configure(d.SSH, d.Config); err != nil {
		return fmt.Errorf("configuring tls: %w", err);
	}
	return nil;

}

func (d *Deployer) runSetupScript() error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting project directory: %w", err)
	}

	localScript := filepath.Join(projectRoot, d.Config.SetupConfig.Script)
	if _, err := os.Stat(localScript); err != nil {
		return fmt.Errorf("setup script %q not found: %w", localScript, err)
	}

	remoteScript := "/tmp/ritta-setup.sh"

	d.log.Infof("Uploading setup script: %s", d.Config.SetupConfig.Script)
	if err := d.SSH.Upload(localScript, remoteScript); err != nil {
		return fmt.Errorf("uploading setup script: %w", err)
	}

	defer func() {
		_ = d.SSH.Run(fmt.Sprintf("rm -f %q", remoteScript))
	}()

	d.log.Info("Running setup script...")
	command := fmt.Sprintf("chmod +x %q", remoteScript)

	if err := d.SSH.Run(command); err != nil {
		return fmt.Errorf("making setup script executable: %w", err)
	}

	if err := d.SSH.RunSudo(remoteScript); err != nil {
		return fmt.Errorf("setup script failed: %w", err)
	}
	return nil
}

func (d *Deployer) prepareSource() error {
	switch d.Config.Source.Type {
	case "existing":
		return d.prepareExistingSource()
	case "git":
		return d.prepareGitSource()
	default:
		return fmt.Errorf("unsupported source type: %q", d.Config.Source.Type)
	}
}

func (d *Deployer) prepareExistingSource() error {
	d.log.Info("Using existing application directory...")

	command := fmt.Sprintf("test -d %q", d.Config.RemoteProjectRoot)

	if err := d.SSH.Run(command); err != nil {
		return fmt.Errorf("application directory does not exist: %w", err)
	}

	return nil
}

func (d *Deployer) prepareGitSource() error {
	root := d.Config.RemoteProjectRoot
	repository := d.Config.Source.Repository
	branch := d.Config.Source.Branch

	d.log.Info("Preparing Git repository...")

	command := fmt.Sprintf(`
	if [ -d %q/.git ]; then
		cd %q &&
		git fetch origin &&
		git checkout %q &&
		git reset --hard origin/%q
	else
		cd ~/%q &&
		git clone --branch %q %q .
	fi
	`, root, root, branch, branch, root, branch, repository)

	return d.SSH.Run(command)
}

func (d *Deployer) build() error {
	if d.Config.Build == nil ||
		d.Config.Build.Command == "" {
		d.log.Info("No build command configured")
		return nil
	}

	d.log.Info("Building application...")
	command := fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, d.Config.Build.Command)

	return d.SSH.Run(command)
}

func (d *Deployer) run() error {
	if d.Config.Run == nil ||
		d.Config.Run.Command == "" {
		return fmt.Errorf("run command is required")
	}

	d.log.Info("Starting application...")
	command := fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, d.Config.Run.Command)

	return d.SSH.Run(command)
}

func (d *Deployer) healthCheck() error {
	if d.Config.Health == nil ||
		d.Config.Health.Command == "" {
		d.log.Info("No health check configured")
		return nil
	}

	d.log.Info("Checking application health...")

	command := fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, d.Config.Health.Command)

	if err := d.SSH.Run(command); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	d.log.Success("Application is healthy :)")

	return nil
}

