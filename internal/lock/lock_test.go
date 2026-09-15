package lock

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeOwner(t *testing.T) {
	tests := []struct {
		name  string
		owner string
		want  string
	}{
		{"trims whitespace", "  alice  ", "alice"},
		{"empty falls back to default", "", "ritta"},
		{"whitespace-only falls back to default", "   ", "ritta"},
		{"strips single quotes", "al'ice", "alice"},
		{"strips newlines", "ali\nce", "alice"},
		{"strips carriage returns", "ali\r\nce", "alice"},
		{"leaves normal names alone", "ci-runner-42", "ci-runner-42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeOwner(tt.owner); got != tt.want {
				t.Errorf("sanitizeOwner(%q) = %q, want %q", tt.owner, got, tt.want)
			}
		})
	}
}

func TestGetLockAcquireScript_SubstitutesAllOwnerOccurrences(t *testing.T) {
	script := getLockAcquireScript("/tmp/ritta-test.lock", "test-owner")

	if strings.Contains(script, "MISSING") {
		t.Fatalf("script contains a missing format argument:\n%s", script)
	}

	const wantOccurrences = 4
	got := strings.Count(script, "'test-owner'")
	if got != wantOccurrences {
		t.Errorf("expected owner to be substituted %d times, got %d\nscript:\n%s", wantOccurrences, got, script)
	}

	if !strings.Contains(script, "LOCK_DIR='/tmp/ritta-test.lock'") {
		t.Errorf("script does not set LOCK_DIR to the given path:\n%s", script)
	}
}

func TestGetLockReleaseScript(t *testing.T) {
	script := getLockReleaseScript("/tmp/ritta-release.lock")

	if strings.Contains(script, "MISSING") {
		t.Fatalf("script contains a missing format argument:\n%s", script)
	}
	if !strings.Contains(script, "LOCK_DIR='/tmp/ritta-release.lock'") {
		t.Errorf("script does not set LOCK_DIR to the given path:\n%s", script)
	}
}

func TestGetLockOwnerScript(t *testing.T) {
	script := getLockOwnerScript("/tmp/ritta-owner.lock")

	if strings.Contains(script, "MISSING") {
		t.Fatalf("script contains a missing format argument:\n%s", script)
	}
	if !strings.Contains(script, "LOCK_DIR='/tmp/ritta-owner.lock'") {
		t.Errorf("script does not set LOCK_DIR to the given path:\n%s", script)
	}
}

type fakeRunner struct {
	calls   []string
	results []error
}

func (f *fakeRunner) Run(command string) error {
	idx := len(f.calls)
	f.calls = append(f.calls, command)
	if idx < len(f.results) {
		return f.results[idx]
	}
	return nil
}

func TestRemoteLock_Acquire_Success(t *testing.T) {
	runner := &fakeRunner{results: []error{nil}}
	l := NewRemoteLock(runner, "/tmp/x.lock").WithOwner("me")

	if err := l.Acquire(); err != nil {
		t.Fatalf("Acquire() = %v, want nil", err)
	}
	if len(runner.calls) != 1 {
		t.Fatalf("expected exactly 1 Run call, got %d", len(runner.calls))
	}
}

func TestRemoteLock_Acquire_HeldByAnotherProcess(t *testing.T) {
	runner := &fakeRunner{results: []error{errors.New("mkdir failed"), nil}}
	l := NewRemoteLock(runner, "/tmp/x.lock").WithOwner("me")

	err := l.Acquire()
	if !errors.Is(err, ErrAlreadyLocked) {
		t.Fatalf("Acquire() = %v, want ErrAlreadyLocked", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 Run calls (acquire + owner check), got %d", len(runner.calls))
	}
}

func TestRemoteLock_Acquire_GenericFailure(t *testing.T) {
	runner := &fakeRunner{results: []error{errors.New("mkdir failed"), errors.New("no info")}}
	l := NewRemoteLock(runner, "/tmp/x.lock").WithOwner("me")

	err := l.Acquire()
	if err == nil {
		t.Fatal("Acquire() = nil, want an error")
	}
	if errors.Is(err, ErrAlreadyLocked) {
		t.Fatalf("Acquire() = %v, want a generic error, not ErrAlreadyLocked", err)
	}
}

func TestRemoteLock_Release_Success(t *testing.T) {
	runner := &fakeRunner{results: []error{nil}}
	l := NewRemoteLock(runner, "/tmp/x.lock")

	if err := l.Release(); err != nil {
		t.Fatalf("Release() = %v, want nil", err)
	}
}

func TestRemoteLock_Release_Failure(t *testing.T) {
	runner := &fakeRunner{results: []error{errors.New("boom")}}
	l := NewRemoteLock(runner, "/tmp/x.lock")

	if err := l.Release(); err == nil {
		t.Fatal("Release() = nil, want an error")
	}
}

func TestRemoteLock_WithOwner_SanitizesBeforeUse(t *testing.T) {
	runner := &fakeRunner{results: []error{errors.New("mkdir failed"), errors.New("no info")}}
	l := NewRemoteLock(runner, "/tmp/x.lock").WithOwner("  ali'ce\n")

	_ = l.Acquire()

	if len(runner.calls) == 0 {
		t.Fatal("expected at least one Run call")
	}
	acquireScript := runner.calls[0]
	if strings.Contains(acquireScript, "'") == false {
		t.Fatalf("expected acquire script to quote the owner, got:\n%s", acquireScript)
	}
	if !strings.Contains(acquireScript, "'alice'") {
		t.Errorf("expected sanitized owner %q in script, got:\n%s", "alice", acquireScript)
	}
	if strings.Contains(acquireScript, "ali'ce") {
		t.Errorf("unsanitized owner leaked into script:\n%s", acquireScript)
	}
}
