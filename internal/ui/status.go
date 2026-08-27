package ui

import (
	"strings"

	"ritta/internal/logger"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type StatusModel struct {
	width  int
	height int

	title string
	logs  []string

	music *MusicPlayer
	log   *logger.Logger

	logCh <-chan logger.Entry
}

type LogMsg struct {
	Entry logger.Entry
}

func NewStatusModel(log *logger.Logger) StatusModel {
	logCh := log.Subscribe()
	mp,_ := NewMusicPlayer(log)
	return StatusModel{
		title:  "RITTA",
		logs:   []string{},
		log:    log,
		logCh:  logCh,
		music: mp,
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
		m.logs = append(m.logs, formatLog(msg.Entry))

		// Keep listening for more logs.
		return *m, waitForLog(m.logCh)
	}

	return *m, nil
}

func formatLog(entry logger.Entry) string {
	time := entry.Time.Format("15:04:05")

	switch entry.Level {
	case logger.Success:
		return time + "  ✓  " + entry.Message

	case logger.Warning:
		return time + "  ⚠  " + entry.Message

	case logger.Error:
		return time + "  ✕  " + entry.Message

	case logger.Debug:
		return time + "  ·  " + entry.Message

	default:
		return time + "  →  " + entry.Message
	}
}

func (m StatusModel) rightView() string {
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

	menu.WriteString(statusTitle.Render("STATUS"))
	menu.WriteString("\n\n")

	menu.WriteString("● READY\n")
	menu.WriteString("  Waiting for operation...")

	menu.WriteString("\n\n")


	menu.WriteString(statusTitle.Render("LOGS"))
	menu.WriteString("\n\n")

	if len(m.logs) == 0 {
		menu.WriteString(
			HomeDimStyle.Render("No logs yet..."),
		)
	} else {
		for _, log := range m.logs {
			menu.WriteString(log)
			menu.WriteString("\n")
		}
	}

	menu.WriteString("\n")


	menu.WriteString(
		HomeDimStyle.Render(
			"────────────────────────────────",
		),
	)

	menu.WriteString("\n\n")

	menu.WriteString(
		HomeDimStyle.Render(
			"'p' Play/Pause Music    'q' Quit",
		),
	)

	return menu.String()
}

func (m StatusModel) handleSelection() (StatusModel, tea.Cmd) {
	return m, nil
}