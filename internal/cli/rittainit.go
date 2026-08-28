package cli

import (
	"fmt"

	"ritta/internal/config"
	"ritta/internal/ui"

	"github.com/spf13/cobra"
)

var configFile string


var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create a new Ritta deployment configuration",

	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := configFile
		if len(args) > 0 {
			targetDir = args[0]
		}
		if err := config.CreateTemplate(targetDir); err != nil {
			return fmt.Errorf("initializing Ritta: %w", err)
		}

		ui.Success("Configuration created successfully")

		fmt.Println()

		fmt.Println(ui.LabelStyle.Render("  Files"))
		fmt.Printf("    %s\n", ui.HomeNormalStyle.Render("rittaConfig.yaml"))
		fmt.Printf("    %s\n", ui.HomeNormalStyle.Render("rittaScript.sh"))

		fmt.Print("\n\n")

		fmt.Println(ui.LabelStyle.Render("  Next steps"))

		fmt.Println("    1. Edit configuration")
		ui.Command("ritta letsconfig rittaConfig.yaml")

		fmt.Println("    2. Validate configuration")
		ui.Command("ritta validate rittaConfig.yaml")

		fmt.Println("    3. Deploy")
		ui.Command("ritta deploy rittaConfig.yaml")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVarP(
		&configFile,
		"path",
		"p",
		"./",
		"Directory for the Ritta deployment configuration",
	)

	initCmd.Flags().StringVarP(
		&configFile,
		"dir",
		"d",
		"./",
		"Directory for the Ritta deployment configuration",
	)
}