package ssh

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)


func (c *Client) Upload(localPath, remotePath string) error {
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return fmt.Errorf("creating SFTP client: %w", err)
	}
	defer sftpClient.Close()

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("opening local file: %w", err)
	}
	defer localFile.Close()

	if err := sftpClient.MkdirAll(filepath.Dir(remotePath)); err != nil {
		return fmt.Errorf("creating remote directory: %w", err)
	}

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("creating remote file: %w", err)
	}
	defer remoteFile.Close()

	if _, err := localFile.WriteTo(remoteFile); err != nil {
		return fmt.Errorf("uploading file: %w", err)
	}

	return nil
}