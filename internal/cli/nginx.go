package cli

import (
	"fmt"
	"os"

	"ritta/internal/config"
	"ritta/internal/logger"
	"ritta/internal/proxy"
	rittaSSH "ritta/internal/ssh"
	"ritta/internal/ui"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var nginxEnv string

var nginxCmd = &cobra.Command{
	Use:   "nginx",
	Short: ui.HomeDescriptionStyle.Render("Manage Nginx reverse proxy configurations"),
}

var nginxSyncCmd = &cobra.Command{
	Use:   "sync [config]",
	Short: ui.HomeDescriptionStyle.Render("Synchronize local deploy/nginx configurations to remote server"),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetFile := "./"
		if len(args) > 0 {
			targetFile = args[0]
		}

		cfg, err := config.LoadConfig(targetFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if err := config.Validate(cfg); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

		log := logger.New(1000)
		log.Infof("Connecting to %s@%s...", cfg.Server.User, cfg.Server.Host)
		client, err := rittaSSH.Connect(cfg.Server.Host, cfg.Server.User, cfg.Server.Key, cfg.Server.Port, log)
		if err != nil {
			log.Errorf("Failed to connect: %v", err)
			return fmt.Errorf("connecting to server: %w", err)
		}
		defer client.Close()
		log.Successf("SSH connected to %s@%s", cfg.Server.User, cfg.Server.Host)

		if cfg.Server.User != "root" {
			fmt.Printf("Sudo password for %s@%s: ", cfg.Server.User, cfg.Server.Host)
			sudoPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("reading sudo password: %w", err)
			}
			sudoPassword := string(sudoPasswordBytes)
			for i := range sudoPasswordBytes {
				sudoPasswordBytes[i] = 0
			}
			client.SetSudoPassword(sudoPassword)
			if err := client.AuthenticateSudo(sudoPassword); err != nil {
				return err
			}
			log.Successf("Sudo authenticated")
		}

		proxyProvider := proxy.NewReverseProxy(client, cfg, log)
		if proxyProvider == nil {
			return fmt.Errorf("nginx proxy provider is not configured")
		}

		if cfg.TLS != nil && cfg.TLS.Provider != "" {
			tlsProvider, err := proxy.NewTLSProvider(client, cfg, log)
			if err == nil {
				if err := tlsProvider.Configure(client, cfg); err != nil {
					log.Warningf("TLS configuration warning: %v", err)
				}
			}
		}

		deployDir := cfg.DeployDir
		if deployDir == "" {
			deployDir = "./deploy"
		}

		if err := proxyProvider.SyncNginx(deployDir, cfg.Domains, nginxEnv); err != nil {
			return fmt.Errorf("syncing nginx: %w", err)
		}

		ui.Success("Nginx configs synchronized and reloaded successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nginxCmd)
	nginxCmd.AddCommand(nginxSyncCmd)

	nginxSyncCmd.Flags().StringVar(
		&nginxEnv,
		"env",
		"all",
		"Target environment for Nginx configs (all, staging, production)",
	)
}
