package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
	"errors"
)

type Client struct {
	client *gossh.Client
}

func Connect(host, user, keyPath string, port int) (*Client, error) {
	keyPath, err := expandHome(keyPath)
	if err != nil {
		return nil, err
	}

	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("reading SSH key: %w", err)
	}

	signer, err := gossh.ParsePrivateKey(key)

	if err != nil {
		var passphraseErr *gossh.PassphraseMissingError

		if errors.As(err, &passphraseErr) {
			fmt.Print("SSH key passphrase: ")

			passphrase, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()

			if err != nil {
				return nil, fmt.Errorf(
					"reading SSH passphrase: %w",
					err,
				)
			}

			signer, err = gossh.ParsePrivateKeyWithPassphrase(
				key,
				passphrase,
			)

			if err != nil {
				return nil, fmt.Errorf(
					"parsing SSH key with passphrase: %w",
					err,
				)
			}
		} else {
			return nil, fmt.Errorf(
				"parsing SSH key: %w",
				err,
			)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf(
			"getting home directory: %w",
			err,
		)
	}

	knownHosts := filepath.Join(
		home,
		".ssh",
		"known_hosts",
	)

	hostKeyCallback, err := knownhosts.New(knownHosts)
	if err != nil {
		return nil, fmt.Errorf(
			"loading known_hosts: %w",
			err,
		)
	}

	config := &gossh.ClientConfig{
		User: user,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
	}

	address := fmt.Sprintf("%s:%d", host, port)

	client, err := gossh.Dial(
		"tcp",
		address,
		config,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"connecting to server: %w",
			err,
		)
	}

	return &Client{
		client: client,
	}, nil
}

func expandHome(path string) (string, error) {
	if path == "~" {
		return os.UserHomeDir();
	}

	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir();
		if err != nil {
			return  "", err;
		}
		return filepath.Join(home, path[2:]), err
	}

	return path, nil
}

func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) Run(command string)  error {
	session, err := c.client.NewSession();
	if err != nil {
		return fmt.Errorf("Error creating session: %w", err);
	}
	defer session.Close();

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Run(command); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (c *Client) Output(command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("creating SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}