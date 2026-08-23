package cli

import (
	"fmt"
	"ritta/internal/config"

	"github.com/spf13/cobra"
)




var validateCmd = &cobra.Command{
	Use: "validate",
	Short: "Validates a deployment configuration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig("deploy.yaml");

		if err != nil {
			return fmt.Errorf("Error loading configuration: %w", err);
		}

		if err = config.Validate(cfg); err != nil {
			return err;
		}

		fmt.Println("Configuration loaded successfully.")
		fmt.Printf("Project location: %s\n", cfg.RootDirectory)
		fmt.Printf("Server: %s@%s\n", cfg.Server.User, cfg.Server.Host)

		return nil;
	},
}

func init() {
	rootCmd.AddCommand(validateCmd);
}