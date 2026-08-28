package cli

import (
	"fmt"
	"ritta/internal/config"
	"ritta/internal/ui"

	"github.com/spf13/cobra"
)


var validateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validates a deployment configuration.",

	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := configFile
		if len(args) > 0 {
			targetPath = args[0]
		}

		cfg, err := config.LoadConfig(targetPath)
		if err != nil {
			return fmt.Errorf("error loading configuration: %w", err)
		}

		if err = config.Validate(cfg); err != nil {
			return err
		}

		ui.Success("Configuration validated successfully.\n")
		ui.Info(fmt.Sprintf("Local project root: %s", cfg.LocalProjectRoot))
		ui.Info(fmt.Sprintf("Remote project location: %s", cfg.RemoteProjectRoot))
		ui.Info(fmt.Sprintf("Source type: %s", cfg.Source.Type))
		ui.Info(fmt.Sprintf("Server: %s@%s", cfg.Server.User, cfg.Server.Host))

		if cfg.Build != nil {
			ui.Info(fmt.Sprintf("Build command: %s", cfg.Build.Command))
		}
		if cfg.Run != nil {
			ui.Info(fmt.Sprintf("Run command: %s", cfg.Run.Command))
		}
		if cfg.Health != nil {
			ui.Info(fmt.Sprintf("Health check command: %s", cfg.Health.Command))
		}
		if cfg.Proxy != nil {
			ui.Info(fmt.Sprintf("Proxy provider: %s", cfg.Proxy.Provider))
		}
		if cfg.TLS != nil {
			ui.Info(fmt.Sprintf("TLS provider: %s", cfg.TLS.Provider))
		}
		ui.Info(fmt.Sprintf("Domains: %v", cfg.Domains))
		ui.Info(fmt.Sprintf("Files: %v", cfg.File))
		ui.Info(fmt.Sprintf("Setup script: %s", cfg.SetupConfig.Script))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVarP(
		&configFile,
		"file",
		"f",
		"./",
		"Path for the Ritta deployment configuration",
	)

	validateCmd.Flags().StringVarP(
		&configFile,
		"path",
		"p",
		"./",
		"Path for the Ritta deployment configuration",
	)
}