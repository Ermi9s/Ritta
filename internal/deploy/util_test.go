package deploy

import (
	"strings"
	"testing"
)

func TestWrapInDir(t *testing.T) {
	got := wrapInDir("/srv/app", "echo hi")
	want := `cd "/srv/app" && echo hi`
	if got != want {
		t.Errorf("wrapInDir() = %q, want %q", got, want)
	}
}

func TestIsDaemonCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"./server &", true},
		{"systemctl start myapp", true},
		{"service myapp start", true},
		{"pm2 start app.js", true},
		{"docker compose up -d", true},
		{"nohup ./server &", true},
		{"./server", false},
		{"npm start", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			if got := isDaemonCommand(tt.cmd); got != tt.want {
				t.Errorf("isDaemonCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestBuildRunCommand(t *testing.T) {
	tests := []struct {
		name   string
		root   string
		rawCmd string
		want   string
	}{
		{
			name:   "daemon command runs as-is",
			root:   "/srv/app",
			rawCmd: "  systemctl start myapp  ",
			want:   `cd "/srv/app" && systemctl start myapp`,
		},
		{
			name:   "plain command gets wrapped in nohup and backgrounded",
			root:   "/srv/app",
			rawCmd: "./server",
			want:   `cd "/srv/app" && (nohup ./server > ritta-app.log 2>&1 &)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRunCommand(tt.root, tt.rawCmd); got != tt.want {
				t.Errorf("buildRunCommand(%q, %q) = %q, want %q", tt.root, tt.rawCmd, got, tt.want)
			}
		})
	}
}

func TestBuildGitSyncCommand(t *testing.T) {
	script := buildGitSyncCommand("/srv/app", "main", "git@github.com:acme/app.git")

	checks := []string{
		`[ -d "/srv/app"/.git ]`,
		`git fetch origin`,
		`git checkout "main"`,
		`git reset --hard origin/"main"`,
		`mkdir -p "/srv/app"`,
		`git clone --branch "main" "git@github.com:acme/app.git" .`,
	}

	for _, want := range checks {
		if !strings.Contains(script, want) {
			t.Errorf("buildGitSyncCommand() missing %q in:\n%s", want, script)
		}
	}
}
