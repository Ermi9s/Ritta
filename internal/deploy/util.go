package deploy

import (
	"fmt"
	"ritta/internal/proxy"
	"os"
	"path/filepath"
)


func (d *Deployer) configureProxy() error {
	if len(d.Config.Domains) == 0 {
		fmt.Println("No domains configured")
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
		fmt.Println("TLS not configured!");
		return nil;
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

	fmt.Printf("Uploading setup script: %s\n", d.Config.SetupConfig.Script)
	if err := d.SSH.Upload(localScript, remoteScript); err != nil {
		return fmt.Errorf("uploading setup script: %w", err)
	}

	defer func() {
		_ = d.SSH.Run(fmt.Sprintf("rm -f %q", remoteScript))
	}()

	fmt.Println("Running setup script...")
	command := fmt.Sprintf("chmod +x %q && %q", remoteScript, remoteScript)

	if err := d.SSH.Run(command); err != nil {
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
	fmt.Println("Using existing application directory...")

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

	fmt.Println("Preparing Git repository...")

	command := fmt.Sprintf(`
	if [ -d %q/.git ]; then
		cd %q &&
		git fetch origin &&
		git checkout %q &&
		git reset --hard origin/%q
	else
		mkdir -p "$(dirname %q)" &&
		git clone --branch %q %q %q
	fi
	`, root, root, branch, branch, root, branch, repository, root)

	return d.SSH.Run(command)
}

func (d *Deployer) build() error {
	if d.Config.Build == nil ||
		d.Config.Build.Command == "" {
		fmt.Println("No build command configured")
		return nil
	}

	fmt.Println("Building application...")
	command := fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, d.Config.Build.Command)

	return d.SSH.Run(command)
}

func (d *Deployer) run() error {
	if d.Config.Run == nil ||
		d.Config.Run.Command == "" {
		return fmt.Errorf("run command is required")
	}

	fmt.Println("Starting application...")
	command := fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, d.Config.Run.Command)

	return d.SSH.Run(command)
}

func (d *Deployer) healthCheck() error {
	if d.Config.Health == nil ||
		d.Config.Health.Command == "" {
		fmt.Println("No health check configured")
		return nil
	}

	fmt.Println("Checking application health...")

	command := fmt.Sprintf("cd %q && %s", d.Config.RemoteProjectRoot, d.Config.Health.Command)

	if err := d.SSH.Run(command); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	fmt.Println(":) Application is healthy")

	return nil
}

