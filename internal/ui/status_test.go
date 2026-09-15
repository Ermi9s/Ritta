package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"ritta/internal/logger"

	tea "charm.land/bubbletea/v2"
)

func TestClampInt(t *testing.T) {
	tests := []struct {
		name        string
		v, min, max int
		want        int
	}{
		{"within range", 5, 0, 10, 5},
		{"below min clamps to min", -3, 0, 10, 0},
		{"above max clamps to max", 15, 0, 10, 10},
		{"equal to min", 0, 0, 10, 0},
		{"equal to max", 10, 0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampInt(tt.v, tt.min, tt.max); got != tt.want {
				t.Errorf("clampInt(%d, %d, %d) = %d, want %d", tt.v, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestMaxScrollOffset(t *testing.T) {
	tests := []struct {
		entryCount int
		want       int
	}{
		{0, 0},
		{1, 0},
		{5, 4},
	}

	for _, tt := range tests {
		if got := maxScrollOffset(tt.entryCount); got != tt.want {
			t.Errorf("maxScrollOffset(%d) = %d, want %d", tt.entryCount, got, tt.want)
		}
	}
}

func makeEntries(n int) []logger.Entry {
	entries := make([]logger.Entry, n)
	for i := range entries {
		entries[i] = logger.Entry{
			Time:    time.Now(),
			Level:   logger.Info,
			Message: fmt.Sprintf("entry-%d", i),
		}
	}
	return entries
}

func TestRightView_PinnedToLatestByDefault(t *testing.T) {
	m := &StatusModel{height: 11, entries: makeEntries(5)}

	view := m.rightView(80)

	for _, want := range []string{"entry-2", "entry-3", "entry-4"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected %q in view, got:\n%s", want, view)
		}
	}
	for _, notWant := range []string{"entry-0", "entry-1"} {
		if strings.Contains(view, notWant) {
			t.Errorf("did not expect %q in view (should have scrolled off), got:\n%s", notWant, view)
		}
	}
	if strings.Contains(view, "scrolled up") {
		t.Errorf("did not expect a scroll indicator when pinned to the live tail, got:\n%s", view)
	}
}

func TestRightView_ScrolledUpShowsOlderEntriesAndIndicator(t *testing.T) {
	m := &StatusModel{height: 11, entries: makeEntries(5), scrollOffset: 2}

	view := m.rightView(80)

	for _, want := range []string{"entry-1", "entry-2"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected %q in view, got:\n%s", want, view)
		}
	}
	for _, notWant := range []string{"entry-0", "entry-3", "entry-4"} {
		if strings.Contains(view, notWant) {
			t.Errorf("did not expect %q in view, got:\n%s", notWant, view)
		}
	}
	if !strings.Contains(view, "scrolled up · 2 newer below") {
		t.Errorf("expected a scroll indicator mentioning 2 newer entries, got:\n%s", view)
	}
}

func TestStatusModel_ScrollKeys(t *testing.T) {
	m := &StatusModel{entries: makeEntries(5)}

	press := func(code rune) {
		t.Helper()
		*m, _ = m.Update(tea.KeyPressMsg{Code: code})
	}

	press(tea.KeyUp)
	if m.scrollOffset != 1 {
		t.Fatalf("after 'up', scrollOffset = %d, want 1", m.scrollOffset)
	}
	press(tea.KeyUp)
	press(tea.KeyUp)
	press(tea.KeyUp)
	press(tea.KeyUp) // one extra press past the max (4 entries back)
	if m.scrollOffset != 4 {
		t.Fatalf("scrollOffset should clamp at maxScrollOffset(5)=4, got %d", m.scrollOffset)
	}

	press(tea.KeyDown)
	if m.scrollOffset != 3 {
		t.Fatalf("after 'down', scrollOffset = %d, want 3", m.scrollOffset)
	}

	press(tea.KeyEnd)
	if m.scrollOffset != 0 {
		t.Fatalf("after 'end', scrollOffset = %d, want 0", m.scrollOffset)
	}

	press(tea.KeyDown) // already at the bottom, must clamp at 0
	if m.scrollOffset != 0 {
		t.Fatalf("scrollOffset should clamp at 0, got %d", m.scrollOffset)
	}

	press(tea.KeyHome)
	if m.scrollOffset != 4 {
		t.Fatalf("after 'home', scrollOffset = %d, want 4", m.scrollOffset)
	}
}
