package cli

import (
	"fmt"

	"ritta/internal/config"

	"github.com/spf13/cobra"
)

var configFile string

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create a new Ritta deployment configuration",

	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.CreateTemplate(configFile); err != nil {
			return fmt.Errorf("initializing Ritta: %w", err)
		}
		
		fmt.Printf(":) Created Ritta configuration: %s\n", configFile)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  Edit manually or run: ritta letsconfig %s\n", configFile)
		fmt.Printf("  Validate:              ritta validate %s\n", configFile)
		fmt.Printf("  Deploy:                ritta deploy %s\n", configFile)

		return nil
	},
}


func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(
		&configFile,
		"file",
		"f",
		config.DefaultConfigFile,
		"Path for the Ritta deployment configuration",
	)

}