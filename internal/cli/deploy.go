package cli

import (
	"fmt"
	"time"
	"ritta/internal/config"
	"ritta/internal/deploy"
	"ritta/internal/logger"
	rittaSSH "ritta/internal/ssh"
	"ritta/internal/ui"
	"os"
	"golang.org/x/term"

	"github.com/spf13/cobra"
)

const actionDelay = 500 * time.Millisecond // just to make things epic
var scanEnv bool

func runDeploy(file string) error {
	log := logger.New(1000)
	time.Sleep(actionDelay)
	cfg, err := config.LoadConfig(file)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	time.Sleep(actionDelay)
	log.Infof("Connecting to %s@%s...", cfg.Server.User, cfg.Server.Host)

	client, err := rittaSSH.Connect(cfg.Server.Host, cfg.Server.User, cfg.Server.Key, cfg.Server.Port, log)
	if err != nil {
		log.Errorf("Failed to connect: %v", err)
		return fmt.Errorf("connecting to server: %w", err)
	}
	defer client.Close()

	time.Sleep(actionDelay)
	log.Successf("SSH connected to %s@%s", cfg.Server.User, cfg.Server.Host)

	fmt.Printf("Sudo password for %s@%s: ", cfg.Server.User, cfg.Server.Host)

	sudoPassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	if err != nil {
		return fmt.Errorf("reading sudo password: %w", err)
	}

	client.SetSudoPassword(string(sudoPassword))


	if err := client.AuthenticateSudo(string(sudoPassword)); err != nil {
		return err
	}

	log.Successf("Sudo authenticated")

	deployer := deploy.New(cfg, client, log)

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
	Short: "Runs the final deployment configuration.",

	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeploy(configFile)
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

}
