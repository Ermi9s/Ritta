package cli

import (
	"ritta/internal/ui"

	"github.com/spf13/cobra"
)



var rootCmd = &cobra.Command{
	Use: "ritta",
	Short: "Ritta deployment tool",
	Long: "Ritta is a deployment tool that you can configure centrally and deploy your applications superrr easily :)",
	Run: func(cmd *cobra.Command, args []string) {
		ui.RunHome()
	},
}

func Execute() error {
	return rootCmd.Execute();
}	

