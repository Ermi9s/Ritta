package cli

import (
	"fmt"
	"os"

	"ritta/internal/lock"
	"ritta/internal/config"
	"ritta/internal/deploy"
	"ritta/internal/logger"
	rittaSSH "ritta/internal/ssh"
	"ritta/internal/ui"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var scanEnv bool
var deployEnv string

func runDeploy(file string) error {
	// setup
	log := logger.New(1000)
	cfg, err := config.LoadConfig(file)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// establish ssh connection
	log.Infof("Connecting to %s@%s...", cfg.Server.User, cfg.Server.Host)
	client, err := rittaSSH.Connect(cfg.Server.Host, cfg.Server.User, cfg.Server.Key, cfg.Server.Port, log)
	if err != nil {
		log.Errorf("Failed to connect: %v", err)
		return fmt.Errorf("connecting to server: %w", err)
	}
	defer client.Close()
	log.Successf("SSH connected to %s@%s", cfg.Server.User, cfg.Server.Host)

	// get sudo password and authenticate if not root
	if cfg.Server.User != "root" {
		fmt.Printf("Sudo password for %s@%s: ", cfg.Server.User, cfg.Server.Host)
		sudoPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("reading sudo password: %w", err)
		}
		sudoPassword := string(sudoPasswordBytes)
		for i := range sudoPasswordBytes {
			sudoPasswordBytes[i] = 0
		}
		client.SetSudoPassword(sudoPassword)
		if err := client.AuthenticateSudo(sudoPassword); err != nil {
			return err
		}
		log.Successf("Sudo authenticated")
	}

	// obtain a lock to prevent concurrent RITTA deployments
	lockPath := fmt.Sprintf("/tmp/ritta-%x.lock", cfg.RemoteProjectRoot)
	deploymentLock := lock.NewRemoteLock(client, lockPath)
	if err := deploymentLock.Acquire(); err != nil {
		return err
	}
	defer func() {
		if err := deploymentLock.Release(); err != nil {
			log.Errorf("Failed to release deployment lock: %v", err)
		}
	}()

	// run deployer 
	deployer := deploy.New(cfg, client, log)
	deployer.SetEnv(deployEnv)
	deployErrCh := make(chan error, 1)
	go func() {
		deployErrCh <- deployer.Deploy(scanEnv)
	}()

	if err := ui.RunApp(log); err != nil {
		return err
	}
	return <-deployErrCh
}

var deployCmd = &cobra.Command{
	Use:   "deploy [path]",
	Short: ui.HomeDescriptionStyle.Render("Runs the final deployment configuration."),
	Long:  ui.HomeDescriptionStyle.Render("Deploys your application to the remote server."),

	RunE: func(cmd *cobra.Command, args []string) error {
		targetFile := configFile
		if len(args) > 0 {
			targetFile = args[0]
		}
		return runDeploy(targetFile)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVarP(
		&configFile,
		"file",
		"f",
		"./",
		"Path for the Ritta deployment configuration",
	)

	deployCmd.Flags().BoolVarP(
		&scanEnv,
		"scan-env",
		"s",
		false,
		"Scan the environment for variables",
	)

	deployCmd.Flags().StringVar(
		&deployEnv,
		"env",
		"all",
		"Target environment for deployment and Nginx configs (all, staging, production)",
	)
}
