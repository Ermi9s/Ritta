package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ritta/internal/logger"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/term"
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

	sshDir := filepath.Join(home, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	if err := ensureKnownHostsFile(sshDir, knownHostsPath); err != nil {
		return nil, err
	}

	hostKeyCallback, err := buildHostKeyCallback(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("loading known_hosts: %w", err)
	}

	config := &gossh.ClientConfig{
		User: user,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
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


func ensureKnownHostsFile(sshDir, knownHostsPath string) error {
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return fmt.Errorf("creating %s: %w", sshDir, err)
	}

	if _, err := os.Stat(knownHostsPath); errors.Is(err, os.ErrNotExist) {
		f, err := os.OpenFile(knownHostsPath, os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("creating %s: %w", knownHostsPath, err)
		}
		f.Close()
	}

	return nil
}

func buildHostKeyCallback(knownHostsPath string) (gossh.HostKeyCallback, error) {
	base, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, err
	}

	return func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		err := base(hostname, remote, key)
		if err == nil {
			return nil
		}

		var keyErr *knownhosts.KeyError
		if !errors.As(err, &keyErr) {
			return err
		}

		if len(keyErr.Want) > 0 {
			return fmt.Errorf(
				"REMOTE HOST IDENTIFICATION HAS CHANGED for %s!\n"+
					"  ssh-keygen -R %s\n"+
					"underlying error: %w",
				hostname, hostname, err,
			)
		}

		return promptAndTrustHostKey(hostname, key, knownHostsPath)
	}, nil
}

func promptAndTrustHostKey(hostname string, key gossh.PublicKey, knownHostsPath string) error {
	fingerprint := gossh.FingerprintSHA256(key)

	fmt.Printf("The authenticity of host '%s' can't be established.\n", hostname)
	fmt.Printf("%s key fingerprint is %s.\n", key.Type(), fingerprint)
	fmt.Print("Are you sure you want to continue connecting (yes/no)? ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading confirmation: %w", err)
	}
	answer = strings.ToLower(strings.TrimSpace(answer))

	if answer != "yes" && answer != "y" {
		return fmt.Errorf("host key for %s not trusted; connection aborted", hostname)
	}

	if err := appendKnownHost(knownHostsPath, hostname, key); err != nil {
		return fmt.Errorf("saving host key: %w", err)
	}

	fmt.Printf("Warning: Permanently added '%s' (%s) to the list of known hosts.\n", hostname, key.Type())
	return nil
}

func appendKnownHost(knownHostsPath, hostname string, key gossh.PublicKey) error {
	normalized := knownhosts.Normalize(hostname)
	line := knownhosts.Line([]string{normalized}, key)

	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(line + "\n"); err != nil {
		return err
	}

	return nil
}

func expandHome(path string) (string, error) {
	if path == "~" {
		return os.UserHomeDir()
	}

	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), err
	}

	return path, nil
}

func (c *Client) SetSudoPassword(password string) {
	c.sudoPassword = password
}

func (c *Client) ClearSudoPassword() {
	c.sudoPassword = ""
}

func (c *Client) Close() {
	c.ClearSudoPassword()
	if c.client != nil {
		c.client.Close()
	}
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

func (c *Client) RunSudoWithStdin(command, stdin string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("error creating session: %w", err)
	}
	defer session.Close()

	// Feed sudo password first, then the actual stdin content
	session.Stdin = strings.NewReader(c.sudoPassword + "\n" + stdin)

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