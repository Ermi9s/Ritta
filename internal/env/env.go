package env

import (
	"fmt"
	"os"
	"path/filepath"

	"ritta/internal/config"
	"ritta/internal/logger"
	rittaSSH "ritta/internal/ssh"
)

type File struct {
	From string
	To   string
}

func Deploy(client *rittaSSH.Client, cfg *config.Config, scanEnv bool, log *logger.Logger) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting project directory: %w", err)
	}

	files := make(map[string]File)

	//automatically discover .env files.
	if scanEnv {
		discovered, err := Scan(projectRoot)
		if err != nil {
			return fmt.Errorf("scanning environment files: %w", err)
		}

		for _, path := range discovered {
			files[path] = File{
				From: path,
				To:   path,
			}
		}
	}

	//explicit files override scanned files.
	for _, file := range cfg.File {
		files[file.From] = File{
			From: file.From,
			To:   file.To,
		}
	}

	if len(files) == 0 {
		log.Info("No environment files configured")
		return nil
	}

	for _, file := range files {
		localPath := filepath.Join(projectRoot, file.From)

		remotePath := file.To
		if !filepath.IsAbs(remotePath) {
			remotePath = filepath.Join(cfg.RemoteProjectRoot, remotePath)
		}

		if _, err := os.Stat(localPath); err != nil {
			return fmt.Errorf("environment file %q not found: %w", localPath, err)
		}

		log.Infof("Uploading %s to remote %s", file.From, remotePath)

		if err := client.Upload(localPath, remotePath); err != nil {
			return fmt.Errorf("uploading %s: %w", file.From, err)
		}
	}

	return nil
}
