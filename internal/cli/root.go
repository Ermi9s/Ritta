package cli

import (
	"ritta/internal/ui"

	"github.com/spf13/cobra"
)

var Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "ritta",
	Short:   ui.HomeDescriptionStyle.Render("Centrally Configured Deployment Tool"),
	Long:    ui.HomeDescriptionStyle.Render("Ritta is a deployment tool that you can configure centrally and deploy your applications superrr easily :)"),
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}
