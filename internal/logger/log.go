package logger

import (
	"fmt"
	"sync"
	"time"
	"bufio"
	"io"
)

type Level int

const (
	Info Level = iota
	Success
	Warning
	Error
	Debug
)

type Entry struct {
	Time    time.Time
	Level   Level
	Message string
}

type Logger struct {
	mu          sync.RWMutex
	entries     []Entry
	subscribers []chan Entry
	maxEntries  int
}


func Pipe(r io.Reader, log *Logger, level Level) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		log.Log(level, line)
	}
}


func New(maxEntries int) *Logger {
	if maxEntries <= 0 {
		maxEntries = 1000
	}

	return &Logger{
		entries:    make([]Entry, 0, maxEntries),
		subscribers: make([]chan Entry, 0),
		maxEntries: maxEntries,
	}
}

func (l *Logger) Log(level Level, message string) {
	entry := Entry{
		Time:    time.Now(),
		Level:   level,
		Message: message,
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, entry)

	if len(l.entries) > l.maxEntries {
		l.entries = l.entries[len(l.entries)-l.maxEntries:]
	}

	for _, subscriber := range l.subscribers {
		select {
		case subscriber <- entry:
		default:
		}
	}
}

func (l *Logger) Info(message string) {
	l.Log(Info, message)
}

func (l *Logger) Success(message string) {
	l.Log(Success, message)
}

func (l *Logger) Warning(message string) {
	l.Log(Warning, message)
}

func (l *Logger) Error(message string) {
	l.Log(Error, message)
}

func (l *Logger) Debug(message string) {
	l.Log(Debug, message)
}

func (l *Logger) Infof(format string, args ...any) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Successf(format string, args ...any) {
	l.Success(fmt.Sprintf(format, args...))
}

func (l *Logger) Warningf(format string, args ...any) {
	l.Warning(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) Debugf(format string, args ...any) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *Logger) Subscribe() <-chan Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	ch := make(chan Entry, 100)

	l.subscribers = append(l.subscribers, ch)

	return ch
}

func (l *Logger) Entries() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	entries := make([]Entry, len(l.entries))
	copy(entries, l.entries)

	return entries
}