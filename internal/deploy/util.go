package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ritta/internal/proxy"
)

func (d *Deployer) configureProxy() error {
	if len(d.Config.Domains) == 0 || d.Config.Proxy == nil || d.Config.Proxy.Provider == "" {
		d.log.Info("No reverse proxy or domains configured")
		return nil
	}

	proxyProvider := proxy.NewReverseProxy(d.SSH, d.Config)
	if proxyProvider == nil {
		return fmt.Errorf("unknown or unsupported proxy provider: %s", d.Config.Proxy.Provider)
	}

	if err := proxyProvider.Configure(d.Config.Domains); err != nil {
		return fmt.Errorf("configuring proxy: %w", err)
	}

	return nil
}

func (d *Deployer) configureTLS() error {
	if d.Config.TLS == nil || d.Config.TLS.Provider == "" {
		d.log.Info("TLS not configured")
		return nil
	}
	tlsprovider, err := proxy.NewTLSProvider(d.SSH, d.Config)
	if err != nil {
		return fmt.Errorf("creating tls provider: %w", err)
	}

	if err := tlsprovider.Configure(d.SSH, d.Config); err != nil {
		return fmt.Errorf("configuring tls: %w", err)
	}
	return nil
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

	//store previous commit hash if available for rollback
	out, err := d.SSH.Output(fmt.Sprintf("cd %q && git rev-parse HEAD 2>/dev/null || true", root))
	if err == nil {
		d.prevCommit = strings.TrimSpace(out)
	}

	command := fmt.Sprintf(`
	if [ -d %q/.git ]; then
		cd %q &&
		git fetch origin &&
		git checkout %q &&
		git reset --hard origin/%q
	else
		mkdir -p %q &&
		cd %q &&
		git clone --branch %q %q .
	fi
	`, root, root, branch, branch, root, root, branch, repository)

	return d.SSH.Run(command)
}

func (d *Deployer) build() error {
	if d.Config.Build == nil || d.Config.Build.Command == "" {
		d.log.Info("No build command configured")
		return nil
	}

	d.log.Info("Building application...")
	command := fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, d.Config.Build.Command)

	return d.SSH.Run(command)
}

func (d *Deployer) run() error {
	if d.Config.Run == nil || d.Config.Run.Command == "" {
		return fmt.Errorf("run command is required")
	}

	d.log.Info("Starting application...")
	runCmd := strings.TrimSpace(d.Config.Run.Command)

	isDaemon := strings.HasSuffix(runCmd, "&") ||
		strings.HasPrefix(runCmd, "systemctl") ||
		strings.HasPrefix(runCmd, "service") ||
		strings.HasPrefix(runCmd, "pm2") ||
		strings.HasPrefix(runCmd, "docker") ||
		strings.HasPrefix(runCmd, "nohup")

	var command string
	if isDaemon {
		command = fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, runCmd)
	} else {
		command = fmt.Sprintf("cd %q && (nohup %s > ritta-app.log 2>&1 &)", d.Config.RemoteProjectRoot, runCmd)
	}

	return d.SSH.Run(command)
}

func (d *Deployer) healthCheck() error {
	if d.Config.Health == nil || d.Config.Health.Command == "" {
		d.log.Info("No health check configured")
		return nil
	}

	d.log.Info("Checking application health...")
	command := fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, d.Config.Health.Command)

	maxAttempts := 15
	pollInterval := 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		d.log.Infof("Running health check (attempt %d/%d)...", attempt, maxAttempts)
		if err := d.SSH.Run(command); err == nil {
			d.log.Success("Application is healthy :)")
			return nil
		}
		if attempt < maxAttempts {
			time.Sleep(pollInterval)
		}
	}

	return fmt.Errorf("health check failed after %d attempts", maxAttempts)
}

func (d *Deployer) rollback() {
	if d.prevCommit == "" || d.Config.Source.Type != "git" {
		return
	}
	d.log.Warningf("Rolling back to previous commit %s...", d.prevCommit)
	rollbackCmd := fmt.Sprintf("cd %q && git checkout %q", d.Config.RemoteProjectRoot, d.prevCommit)
	if err := d.SSH.Run(rollbackCmd); err != nil {
		d.log.Errorf("Rollback failed: %v", err)
		return
	}
	if d.Config.Build != nil && d.Config.Build.Command != "" {
		_ = d.build()
	}
	if d.Config.Run != nil && d.Config.Run.Command != "" {
		_ = d.run()
	}
	d.log.Success("Rollback completed")
}
