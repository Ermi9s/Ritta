package env

import (
	"fmt"
	"os"
	"path/filepath"

	"ritta/internal/config"
	rittaSSH "ritta/internal/ssh"
)

type EnvFile struct {
	From string
	To   string
}

func Deploy(client *rittaSSH.Client, cfg *config.Config) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting project directory: %w", err)
	}

	files := make(map[string]EnvFile)

	// Automatically discover .env files.
	if cfg.Env.Scan {
		discovered, err := Scan(projectRoot)
		if err != nil {
			return fmt.Errorf("scanning environment files: %w", err)
		}

		for _, path := range discovered {
			files[path] = EnvFile{
				From: path,
				To:   path,
			}
		}
	}

	// Explicit files override scanned files.
	for _, file := range cfg.Env.Files {
		files[file.From] = EnvFile{
			From: file.From,
			To:   file.To,
		}
	}

	if len(files) == 0 {
		fmt.Println("No environment files configured")
		return nil
	}

	for _, file := range files {
		localPath := filepath.Join(
			projectRoot,
			file.From,
		)

		remotePath := filepath.Join(
			cfg.RootDirectory,
			file.To,
		)

		if _, err := os.Stat(localPath); err != nil {
			return fmt.Errorf("environment file %q not found: %w", localPath, err)
		}

		fmt.Printf("Uploading %s to remote %s\n", file.From, remotePath)

		if err := client.Upload(localPath, remotePath); err != nil {
			return fmt.Errorf("uploading %s: %w", file.From, err)
		}
	}

	return nil
}