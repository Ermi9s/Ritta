package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)



var rootCmd = &cobra.Command{
	Use: "ritta",
	Short: "Ritta deployment tool",
	Long: "Ritta is a deployment tool that you can configure centrally and deploy your applications easily :)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Ritta ;)")
	},
}

func Execute() error {
	return rootCmd.Execute();
}	

