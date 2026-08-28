package ui

import (
	"ritta/internal/logger"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type StatusModel struct {
	width  int
	height int

	title   string
	entries []logger.Entry

	music *MusicPlayer
	log   *logger.Logger

	logCh <-chan logger.Entry
}

type LogMsg struct {
	Entry logger.Entry
}

func NewStatusModel(log *logger.Logger) StatusModel {
	logCh := log.Subscribe()
	mp, _ := NewMusicPlayer(log)
	mp.Play()

	log.Success("Kick back, playing The astounding eyes of rita . . . :))))")

	return StatusModel{
		title:   "RITTA",
		entries: []logger.Entry{},
		log:     log,
		logCh:   logCh,
		music:   mp,
	}
}

func (m StatusModel) Init() tea.Cmd {
	if m.log == nil {
		return nil
	}

	m.logCh = m.log.Subscribe()

	return waitForLog(m.logCh)
}

func waitForLog(ch <-chan logger.Entry) tea.Cmd {
	return func() tea.Msg {
		entry := <-ch

		return LogMsg{
			Entry: entry,
		}
	}
}

func (m *StatusModel) Update(msg tea.Msg) (StatusModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {

		case "p":
			if m.music != nil {
				if err := m.music.Toggle(); err != nil {
					m.log.Error(err.Error())
				}
			}

		case "q", "ctrl+c":
			if m.music != nil {
				_ = m.music.Close()
			}
			return *m, tea.Quit
		}

	case LogMsg:
		m.entries = append(m.entries, msg.Entry)

		return *m, waitForLog(m.logCh)
	}

	return *m, nil
}

// logLevelStyle returns the icon and lipgloss style for a given log level.
func logLevelStyle(level logger.Level) (string, lipgloss.Style) {
	switch level {
	case logger.Success:
		return "✓", lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80"))
	case logger.Warning:
		return "⚠", lipgloss.NewStyle().Foreground(lipgloss.Color("#FACC15"))
	case logger.Error:
		return "✕", lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F87171"))
	case logger.Debug:
		return "·", lipgloss.NewStyle().Foreground(lipgloss.Color("#71717A"))
	default:
		return "→", lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8"))
	}
}

func renderEntry(entry logger.Entry, maxWidth int) string {
	timestamp := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#52525B")).
		Render(entry.Time.Format("15:04:05"))

	icon, style := logLevelStyle(entry.Level)
	iconStr := style.Render(icon)

	const prefixLen = 13
	msgWidth := maxWidth - prefixLen
	if msgWidth < 20 {
		msgWidth = 20
	}

	words := strings.Fields(entry.Message)
	var lines []string
	var current strings.Builder

	for _, word := range words {
		if current.Len() == 0 {
			current.WriteString(word)
		} else if current.Len()+1+len(word) <= msgWidth {
			current.WriteByte(' ')
			current.WriteString(word)
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	if len(lines) == 0 {
		lines = []string{""}
	}

	indent := strings.Repeat(" ", prefixLen)
	var sb strings.Builder
	for i, line := range lines {
		if i == 0 {
			sb.WriteString(timestamp)
			sb.WriteString("  ")
			sb.WriteString(iconStr)
			sb.WriteString("  ")
			sb.WriteString(style.Render(line))
		} else {
			sb.WriteByte('\n')
			sb.WriteString(indent)
			sb.WriteString(style.Render(line))
		}
	}
	return sb.String()
}

func (m *StatusModel) rightView(contentWidth int) string {
	var menu strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#B8A1FF")).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("#376E88")).
		Padding(0, 1)

	menu.WriteString(headerStyle.Render(m.title))
	menu.WriteString("\n\n")

	statusTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#B8A1FF"))

	menu.WriteString(statusTitle.Render("LOGS"))
	menu.WriteString("\n\n")

	if len(m.entries) == 0 {
		menu.WriteString(HomeDimStyle.Render("No logs yet..."))
	} else {
		const reservedLines = 8
		maxLogLines := m.height - reservedLines
		if maxLogLines < 1 {
			maxLogLines = 1
		}

		var visible []string
		usedLines := 0
		for i := len(m.entries) - 1; i >= 0; i-- {
			rendered := renderEntry(m.entries[i], contentWidth)
			lines := lipgloss.Height(rendered)

			if usedLines+lines > maxLogLines {
				break
			}

			visible = append(visible, rendered)
			usedLines += lines
		}

		for i, j := 0, len(visible)-1; i < j; i, j = i+1, j-1 {
			visible[i], visible[j] = visible[j], visible[i]
		}

		menu.WriteString(strings.Join(visible, "\n"))
	}

	menu.WriteString("\n\n")
	menu.WriteString(
		HomeDimStyle.Render("────────────────────────────────"),
	)
	menu.WriteString("\n\n")
	menu.WriteString(
		HomeDimStyle.Render("'p' Play/Pause Music    'q' Quit"),
	)

	return menu.String()
}

func (m StatusModel) handleSelection() (StatusModel, tea.Cmd) {
	return m, nil
}
