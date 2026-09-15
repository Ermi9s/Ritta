package lock

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrAlreadyLocked = errors.New("deployment already in progress")
	ErrInvalidLock   = errors.New("invalid deployment lock")
)

type RemoteRunner interface {
	Run(command string) error
}

type RemoteLock struct {
	client RemoteRunner
	path   string
	owner  string
}

func NewRemoteLock(client RemoteRunner, path string) *RemoteLock {
	return &RemoteLock{
		client: client,
		path:   path,
	}
}

func (l *RemoteLock) Acquire() error {
	
	script := getLockAcquireScript(l.path, l.owner)

	err := l.client.Run(script)
	if err == nil {
		return nil
	}

	if l.isHeld() {
		return fmt.Errorf("%w: %s", ErrAlreadyLocked, l.path)
	}

	return fmt.Errorf("acquiring remote lock: %w", err)
}


func (l *RemoteLock) Release() error {
	script := getLockReleaseScript(l.path)
	if err := l.client.Run(script); err != nil {
		return fmt.Errorf("releasing remote lock: %w", err)
	}

	return nil
}

func (l *RemoteLock) isHeld() bool {
	script := getLockOwnerScript(l.path)
	return l.client.Run(script) == nil
}


func (l *RemoteLock) WithOwner(owner string) *RemoteLock {
	l.owner = sanitizeOwner(owner)
	return l
}

func sanitizeOwner(owner string) string {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return "ritta"
	}
	return strings.NewReplacer(
		"'", "",
		"\n", "",
		"\r", "",
	).Replace(owner)
}

