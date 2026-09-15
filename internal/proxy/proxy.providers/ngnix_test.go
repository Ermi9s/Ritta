package proxyproviders

import (
	"errors"
	"strings"
	"testing"

	"ritta/internal/config"
	"ritta/internal/logger"
)

func testLogger() *logger.Logger {
	return logger.New(100)
}

type fakeSSHRunner struct {
	sudoCalls       []string
	sudoStdinCalls  []string
	sudoStdinInputs []string
	results         []error
}

func (f *fakeSSHRunner) next() error {
	idx := len(f.sudoCalls) + len(f.sudoStdinCalls) - 1
	if idx >= 0 && idx < len(f.results) {
		return f.results[idx]
	}
	return nil
}

func (f *fakeSSHRunner) RunSudo(command string) error {
	f.sudoCalls = append(f.sudoCalls, command)
	return f.next()
}

func (f *fakeSSHRunner) RunSudoWithStdin(command, stdin string) error {
	f.sudoStdinCalls = append(f.sudoStdinCalls, command)
	f.sudoStdinInputs = append(f.sudoStdinInputs, stdin)
	return f.next()
}

func TestGenerateConfig(t *testing.T) {
	domain := config.Domain{Host: "something.com", Port: 8080}
	conf := GenerateConfig(domain)

	checks := []string{
		"listen 80;",
		"listen [::]:80;",
		"server_name something.com;",
		"proxy_pass http://127.0.0.1:8080;",
		"proxy_set_header Host $host;",
		"proxy_set_header X-Real-IP $remote_addr;",
		"proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;",
		"proxy_set_header X-Forwarded-Proto $scheme;",
	}
	for _, want := range checks {
		if !strings.Contains(conf, want) {
			t.Errorf("GenerateConfig() missing %q in:\n%s", want, conf)
		}
	}
}

func TestNginx_EnsureInstalled(t *testing.T) {
	t.Run("nginx present", func(t *testing.T) {
		runner := &fakeSSHRunner{}
		n := NewNginx(runner, testLogger())
		if err := n.EnsureInstalled(); err != nil {
			t.Errorf("EnsureInstalled() = %v, want nil", err)
		}
	})

	t.Run("nginx missing", func(t *testing.T) {
		runner := &fakeSSHRunner{results: []error{errors.New("not found")}}
		n := NewNginx(runner, testLogger())
		if err := n.EnsureInstalled(); err == nil {
			t.Error("EnsureInstalled() = nil, want an error")
		}
	})
}

func TestNginx_ConfigureDomain(t *testing.T) {
	runner := &fakeSSHRunner{}
	n := NewNginx(runner, testLogger())
	domain := config.Domain{Host: "something.com", Port: 3000}

	if err := n.ConfigureDomain(domain); err != nil {
		t.Fatalf("ConfigureDomain() = %v, want nil", err)
	}

	if len(runner.sudoStdinCalls) != 1 {
		t.Fatalf("expected 1 RunSudoWithStdin call, got %d", len(runner.sudoStdinCalls))
	}
	if !strings.Contains(runner.sudoStdinCalls[0], "/etc/nginx/conf.d/ritta-something.com.conf") {
		t.Errorf("expected write to ritta-something.com.conf, got command: %q", runner.sudoStdinCalls[0])
	}
	if !strings.Contains(runner.sudoStdinInputs[0], "server_name something.com;") {
		t.Errorf("expected config content to target something.com, got:\n%s", runner.sudoStdinInputs[0])
	}

	t.Run("write failure propagates", func(t *testing.T) {
		runner := &fakeSSHRunner{results: []error{errors.New("permission denied")}}
		n := NewNginx(runner, testLogger())
		if err := n.ConfigureDomain(domain); err == nil {
			t.Error("ConfigureDomain() = nil, want an error")
		}
	})
}

func TestNginx_Test(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		runner := &fakeSSHRunner{}
		n := NewNginx(runner, testLogger())
		if err := n.Test(); err != nil {
			t.Errorf("Test() = %v, want nil", err)
		}
		if len(runner.sudoCalls) != 1 || runner.sudoCalls[0] != "nginx -t" {
			t.Errorf("expected a single 'nginx -t' call, got %v", runner.sudoCalls)
		}
	})

	t.Run("fails", func(t *testing.T) {
		runner := &fakeSSHRunner{results: []error{errors.New("syntax error")}}
		n := NewNginx(runner, testLogger())
		if err := n.Test(); err == nil {
			t.Error("Test() = nil, want an error")
		}
	})
}

func TestNginx_Reload(t *testing.T) {
	t.Run("enables then reloads", func(t *testing.T) {
		runner := &fakeSSHRunner{}
		n := NewNginx(runner, testLogger())
		if err := n.Reload(); err != nil {
			t.Fatalf("Reload() = %v, want nil", err)
		}
		want := []string{"systemctl enable --now nginx", "systemctl reload nginx"}
		if len(runner.sudoCalls) != len(want) {
			t.Fatalf("expected calls %v, got %v", want, runner.sudoCalls)
		}
		for i, w := range want {
			if runner.sudoCalls[i] != w {
				t.Errorf("call %d = %q, want %q", i, runner.sudoCalls[i], w)
			}
		}
	})

	t.Run("stops if enable fails", func(t *testing.T) {
		runner := &fakeSSHRunner{results: []error{errors.New("enable failed")}}
		n := NewNginx(runner, testLogger())
		if err := n.Reload(); err == nil {
			t.Fatal("Reload() = nil, want an error")
		}
		if len(runner.sudoCalls) != 1 {
			t.Errorf("expected Reload to stop after enable failure, got calls: %v", runner.sudoCalls)
		}
	})
}

func TestNginx_Configure(t *testing.T) {
	t.Run("no domains is a no-op success", func(t *testing.T) {
		runner := &fakeSSHRunner{}
		n := NewNginx(runner, testLogger())
		if err := n.Configure(nil); err != nil {
			t.Errorf("Configure(nil) = %v, want nil", err)
		}
		if len(runner.sudoCalls) != 0 || len(runner.sudoStdinCalls) != 0 {
			t.Errorf("expected no SSH calls for empty domain list, got sudo=%v stdin=%v", runner.sudoCalls, runner.sudoStdinCalls)
		}
	})

	t.Run("full happy path runs install, configure, test, reload in order", func(t *testing.T) {
		runner := &fakeSSHRunner{}
		n := NewNginx(runner, testLogger())
		domains := []config.Domain{
			{Host: "a.something.com", Port: 3000},
			{Host: "b.something.com", Port: 3001},
		}

		if err := n.Configure(domains); err != nil {
			t.Fatalf("Configure() = %v, want nil", err)
		}

		if len(runner.sudoStdinCalls) != 2 {
			t.Errorf("expected a config write per domain, got %d", len(runner.sudoStdinCalls))
		}

		wantSudo := []string{
			"command -v nginx >/dev/null 2>&1 || [ -x /usr/sbin/nginx ]",
			"nginx -t",
			"systemctl enable --now nginx",
			"systemctl reload nginx",
		}
		if len(runner.sudoCalls) != len(wantSudo) {
			t.Fatalf("RunSudo calls = %v, want %v", runner.sudoCalls, wantSudo)
		}
		for i, w := range wantSudo {
			if runner.sudoCalls[i] != w {
				t.Errorf("RunSudo call %d = %q, want %q", i, runner.sudoCalls[i], w)
			}
		}
	})

	t.Run("stops if a domain fails to configure", func(t *testing.T) {
		runner := &fakeSSHRunner{results: []error{nil, errors.New("disk full")}}
		n := NewNginx(runner, testLogger())
		domains := []config.Domain{{Host: "a.something.com", Port: 3000}}

		if err := n.Configure(domains); err == nil {
			t.Fatal("Configure() = nil, want an error")
		}
		if len(runner.sudoCalls) != 1 {
			t.Errorf("expected Configure to stop after the failed domain (no Test/Reload), got sudo calls: %v", runner.sudoCalls)
		}
	})
}
