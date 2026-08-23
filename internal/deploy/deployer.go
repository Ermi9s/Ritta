package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"ritta/internal/config"
	"ritta/internal/env"
	"ritta/internal/proxy"
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

func (d *Deployer) configureProxy() error {
	if len(d.Config.Domains) == 0 {
		fmt.Println("No domains configured")
		return nil
	}

	nginx := proxy.NewNginx(d.SSH)

	if err := nginx.Configure(d.Config.Domains); err != nil {
		return err
	}

	if err := proxy.ConfigureTLS(
		d.SSH,
		d.Config,
	); err != nil {
		return err
	}

	return nil
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
		return fmt.Errorf(
			"configuring reverse proxy: %w",
			err,
		)
	}

	fmt.Println()
	fmt.Println(":) Deployment complete")

	return nil
}

func (d *Deployer) runSetupScript() error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting project directory: %w", err)
	}

	localScript := filepath.Join(
		projectRoot,
		d.Config.SetupConfig.Script,
	)

	if _, err := os.Stat(localScript); err != nil {
		return fmt.Errorf(
			"setup script %q not found: %w",
			localScript,
			err,
		)
	}

	remoteScript := "/tmp/ritta-setup.sh"

	fmt.Printf(
		"Uploading setup script: %s\n",
		d.Config.SetupConfig.Script,
	)

	if err := d.SSH.Upload(
		localScript,
		remoteScript,
	); err != nil {
		return fmt.Errorf(
			"uploading setup script: %w",
			err,
		)
	}

	defer func() {
		_ = d.SSH.Run(
			fmt.Sprintf("rm -f %q", remoteScript),
		)
	}()

	fmt.Println("Running setup script...")

	command := fmt.Sprintf(
		"chmod +x %q && %q",
		remoteScript,
		remoteScript,
	)

	if err := d.SSH.Run(command); err != nil {
		return fmt.Errorf(
			"setup script failed: %w",
			err,
		)
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
		return fmt.Errorf(
			"unsupported source type: %q",
			d.Config.Source.Type,
		)
	}
}

func (d *Deployer) prepareExistingSource() error {
	fmt.Println("Using existing application directory...")

	command := fmt.Sprintf(
		"test -d %q",
		d.Config.RootDirectory,
	)

	if err := d.SSH.Run(command); err != nil {
		return fmt.Errorf(
			"application directory does not exist: %w",
			err,
		)
	}

	return nil
}

func (d *Deployer) prepareGitSource() error {
	root := d.Config.RootDirectory
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

	command := fmt.Sprintf(
		"cd %q && %s",
		d.Config.RootDirectory,
		d.Config.Build.Command,
	)

	return d.SSH.Run(command)
}

func (d *Deployer) run() error {
	if d.Config.Run == nil ||
		d.Config.Run.Command == "" {
		return fmt.Errorf("run command is required")
	}

	fmt.Println("Starting application...")

	command := fmt.Sprintf(
		"cd %q && %s",
		d.Config.RootDirectory,
		d.Config.Run.Command,
	)

	return d.SSH.Run(command)
}

func (d *Deployer) healthCheck() error {
	if d.Config.Health == nil ||
		d.Config.Health.Command == "" {
		fmt.Println("No health check configured")
		return nil
	}

	fmt.Println("Checking application health...")

	command := fmt.Sprintf(
		"cd %q && %s",
		d.Config.RootDirectory,
		d.Config.Health.Command,
	)

	if err := d.SSH.Run(command); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	fmt.Println(":) Application is healthy")

	return nil
}