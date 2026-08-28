package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"ritta/internal/logger"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
	"errors"
	"io"
	"strings"
)

type Client struct {
	client       *gossh.Client
	log          *logger.Logger
	sudoPassword string
}


func Connect(host, user, keyPath string, port int, log *logger.Logger) (*Client, error) {
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
		log:    log,
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

func (c *Client) SetSudoPassword(password string) {
	c.sudoPassword = password
}


func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) Run(command string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("Error creating session: %w", err)
	}
	defer session.Close()

	if c.log != nil {
		// Pipe stdout and stderr through the logger so they appear in the UI
		stdoutR, stdoutW := io.Pipe()
		stderrR, stderrW := io.Pipe()
		session.Stdout = stdoutW
		session.Stderr = stderrW

		done := make(chan struct{}, 2)
		go func() { logger.Pipe(stdoutR, c.log, logger.Info); done <- struct{}{} }()
		go func() { logger.Pipe(stderrR, c.log, logger.Info); done <- struct{}{} }()

		runErr := session.Run(command)
		stdoutW.Close()
		stderrW.Close()
		<-done
		<-done

		if runErr != nil {
			return fmt.Errorf("command failed: %w", runErr)
		}
	} else {
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
		if err := session.Run(command); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	return nil
}

func (c *Client) RunSudo(command string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}
	defer session.Close()

	session.Stdin = strings.NewReader(c.sudoPassword + "\n")

	if c.log != nil {
		stdoutR, stdoutW := io.Pipe()
		stderrR, stderrW := io.Pipe()

		session.Stdout = stdoutW
		session.Stderr = stderrW

		done := make(chan struct{}, 2)

		go func() {
			logger.Pipe(stdoutR, c.log, logger.Info)
			done <- struct{}{}
		}()

		go func() {
			logger.Pipe(stderrR, c.log, logger.Info)
			done <- struct{}{}
		}()

		runErr := session.Run("sudo -S " + command)

		stdoutW.Close()
		stderrW.Close()

		<-done
		<-done

		if runErr != nil {
			return fmt.Errorf("command failed: %w", runErr)
		}

		return nil
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Run("sudo -S " + command); err != nil {
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

func (c *Client) AuthenticateSudo(password string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("creating sudo session: %w", err)
	}
	defer session.Close()

	session.Stdin = strings.NewReader(password + "\n")

	output, err := session.CombinedOutput("sudo -S -v")
	if err != nil {
		return fmt.Errorf("sudo authentication failed: %w: %s", err, output)
	}

	return nil
}