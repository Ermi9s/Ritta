package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() error: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"bare tilde expands to home", "~", home},
		{"tilde slash expands and joins", "~/.ssh/id_ed25519", filepath.Join(home, ".ssh/id_ed25519")},
		{"absolute path is unchanged", "/etc/ritta/key", "/etc/ritta/key"},
		{"relative path is unchanged", "keys/id_ed25519", "keys/id_ed25519"},
		{"tilde without slash is left alone", "~ermias", "~ermias"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandHome(tt.path)
			if err != nil {
				t.Fatalf("expandHome(%q) error: %v", tt.path, err)
			}
			if got != tt.want {
				t.Errorf("expandHome(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestEnsureKnownHostsFile(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	knownHostsPath := filepath.Join(sshDir, "known_hosts")

	if err := ensureKnownHostsFile(sshDir, knownHostsPath); err != nil {
		t.Fatalf("ensureKnownHostsFile() error: %v", err)
	}

	info, err := os.Stat(knownHostsPath)
	if err != nil {
		t.Fatalf("known_hosts file was not created: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("known_hosts permissions = %v, want 0600", info.Mode().Perm())
	}

	// Calling it again on an already-existing file must not error or wipe it.
	if err := os.WriteFile(knownHostsPath, []byte("existing-content\n"), 0o600); err != nil {
		t.Fatalf("seeding known_hosts failed: %v", err)
	}
	if err := ensureKnownHostsFile(sshDir, knownHostsPath); err != nil {
		t.Fatalf("ensureKnownHostsFile() on existing file error: %v", err)
	}
	content, err := os.ReadFile(knownHostsPath)
	if err != nil {
		t.Fatalf("reading known_hosts failed: %v", err)
	}
	if string(content) != "existing-content\n" {
		t.Errorf("ensureKnownHostsFile clobbered existing content: %q", content)
	}
}

func generateTestKey(t *testing.T) gossh.PublicKey {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}
	sshPub, err := gossh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("converting to ssh public key: %v", err)
	}
	return sshPub
}

func seedKnownHosts(t *testing.T, path, hostname string, key gossh.PublicKey) {
	t.Helper()
	line := knownhosts.Line([]string{knownhosts.Normalize(hostname)}, key)
	if err := os.WriteFile(path, []byte(line+"\n"), 0o600); err != nil {
		t.Fatalf("seeding known_hosts: %v", err)
	}
}

func TestBuildHostKeyCallback_KnownHostMatches(t *testing.T) {
	dir := t.TempDir()
	knownHostsPath := filepath.Join(dir, "known_hosts")
	key := generateTestKey(t)
	seedKnownHosts(t, knownHostsPath, "example.com:22", key)

	callback, err := buildHostKeyCallback(knownHostsPath)
	if err != nil {
		t.Fatalf("buildHostKeyCallback() error: %v", err)
	}

	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	if err := callback("example.com:22", addr, key); err != nil {
		t.Errorf("callback() with matching known key = %v, want nil", err)
	}
}

func TestBuildHostKeyCallback_ChangedHostKeyIsRejected(t *testing.T) {
	dir := t.TempDir()
	knownHostsPath := filepath.Join(dir, "known_hosts")
	trustedKey := generateTestKey(t)
	attackerKey := generateTestKey(t)
	seedKnownHosts(t, knownHostsPath, "example.com:22", trustedKey)

	callback, err := buildHostKeyCallback(knownHostsPath)
	if err != nil {
		t.Fatalf("buildHostKeyCallback() error: %v", err)
	}

	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	err = callback("example.com:22", addr, attackerKey)
	if err == nil {
		t.Fatal("callback() with changed host key = nil, want an error")
	}
	if !strings.Contains(err.Error(), "REMOTE HOST IDENTIFICATION HAS CHANGED") {
		t.Errorf("callback() error = %v, want a host-identification-changed warning", err)
	}
}

func TestBuildHostKeyCallback_UnknownHostPromptsAndTrustsOnYes(t *testing.T) {
	dir := t.TempDir()
	knownHostsPath := filepath.Join(dir, "known_hosts")
	if err := os.WriteFile(knownHostsPath, nil, 0o600); err != nil {
		t.Fatalf("creating empty known_hosts: %v", err)
	}
	key := generateTestKey(t)

	callback, err := buildHostKeyCallback(knownHostsPath)
	if err != nil {
		t.Fatalf("buildHostKeyCallback() error: %v", err)
	}

	restoreStdin := simulateStdin(t, "yes\n")
	defer restoreStdin()

	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	if err := callback("newhost.com:22", addr, key); err != nil {
		t.Fatalf("callback() for unknown host with 'yes' answer = %v, want nil", err)
	}

	content, err := os.ReadFile(knownHostsPath)
	if err != nil {
		t.Fatalf("reading known_hosts: %v", err)
	}
	if !strings.Contains(string(content), "newhost.com") {
		t.Errorf("expected known_hosts to now contain newhost.com, got:\n%s", content)
	}
}

func TestBuildHostKeyCallback_UnknownHostRejectedOnNo(t *testing.T) {
	dir := t.TempDir()
	knownHostsPath := filepath.Join(dir, "known_hosts")
	if err := os.WriteFile(knownHostsPath, nil, 0o600); err != nil {
		t.Fatalf("creating empty known_hosts: %v", err)
	}
	key := generateTestKey(t)

	callback, err := buildHostKeyCallback(knownHostsPath)
	if err != nil {
		t.Fatalf("buildHostKeyCallback() error: %v", err)
	}

	restoreStdin := simulateStdin(t, "no\n")
	defer restoreStdin()

	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	if err := callback("newhost.com:22", addr, key); err == nil {
		t.Fatal("callback() for unknown host with 'no' answer = nil, want an error")
	}

	content, err := os.ReadFile(knownHostsPath)
	if err != nil {
		t.Fatalf("reading known_hosts: %v", err)
	}
	if strings.Contains(string(content), "newhost.com") {
		t.Errorf("known_hosts should not have been modified after rejecting the host, got:\n%s", content)
	}
}

// simulateStdin temporarily replaces os.Stdin with a pipe pre-loaded with
// the given input, and returns a function that restores the original.
func simulateStdin(t *testing.T, input string) func() {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdin pipe: %v", err)
	}
	if _, err := io.WriteString(w, input); err != nil {
		t.Fatalf("writing simulated stdin: %v", err)
	}
	w.Close()

	original := os.Stdin
	os.Stdin = r
	return func() {
		os.Stdin = original
		r.Close()
	}
}
