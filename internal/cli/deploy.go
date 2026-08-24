package cli

import (
	"fmt"
	"ritta/internal/config"
	"ritta/internal/deploy"
	rittaSSH "ritta/internal/ssh"

	"github.com/spf13/cobra"
)


func runDeploy(file string) error {
	cfg, err := config.LoadConfig(file)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	fmt.Printf("Connecting to %s@%s...\n", cfg.Server.User, cfg.Server.Host)

	client, err := rittaSSH.Connect(cfg.Server.Host, cfg.Server.User,cfg.Server.Key, cfg.Server.Port);

	if err != nil {
		return fmt.Errorf("connecting to server: %w", err)
	}

	defer client.Close()

	fmt.Println("SSH connected :)")

	deployer := deploy.New(cfg, client)

	return deployer.Deploy()
}

var deployCmd = &cobra.Command{
	Use: "deploy [path]",
	Short: "Runs the final deployment configuration.",

	RunE: func(cmd *cobra.Command, args []string) error {
		return runDeploy(configFile);
	},
}


func init() {
	rootCmd.AddCommand(deployCmd);
	deployCmd.Flags().StringVarP(
		&configFile,
		"file",
		"f",
		config.DefaultConfigFile,
		"Path for the Ritta deployment configuration",
	)

}