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
	Short: ui.HomeDescriptionStyle.Render("	Create a new Ritta deployment configuration"),
	Long:  ui.HomeDescriptionStyle.Render("Initializes a new Ritta deployment configuration. This command creates a new rittaConfig.yaml, rittaScript.sh, and local deploy/ directory with standardized Nginx configs."),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := configFile
		if len(args) > 0 {
			targetDir = args[0]
		}
		if err := config.CreateTemplate(targetDir); err != nil {
			return fmt.Errorf("initializing Ritta: %w", err)
		}

		ui.Success("Configuration and deploy directory created successfully")

		fmt.Println()

		fmt.Println(ui.LabelStyle.Render("  Files & Directories"))
		fmt.Printf("    %s\n", ui.HomeNormalStyle.Render("rittaConfig.yaml"))
		fmt.Printf("    %s\n", ui.HomeNormalStyle.Render("rittaScript.sh"))
		fmt.Printf("    %s\n", ui.HomeNormalStyle.Render("deploy/nginx/conf.d/global.conf"))
		fmt.Printf("    %s\n", ui.HomeNormalStyle.Render("deploy/nginx/sites-available/app.example.com.conf"))
		fmt.Printf("    %s\n", ui.HomeNormalStyle.Render("deploy/.env.deploy.example"))

		fmt.Print("\n\n")

		fmt.Println(ui.LabelStyle.Render("  Next steps"))

		fmt.Println("    1. Configure deployment environment credentials")
		ui.Command("cp deploy/.env.deploy.example .env.deploy")

		fmt.Println("    2. Edit configuration")
		ui.Command("ritta letsconfig rittaConfig.yaml")

		fmt.Println("    3. Validate configuration")
		ui.Command("ritta validate rittaConfig.yaml")

		fmt.Println("    4. Deploy")
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
