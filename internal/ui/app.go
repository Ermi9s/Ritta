package ui

import (
	"ritta/internal/logger"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Screen int

const (
	StatusScreen Screen = iota
)

type AppModel struct {
	screen Screen
	status StatusModel
}


func NewApp(log *logger.Logger) AppModel {
	return AppModel{
		screen: StatusScreen,
		status: NewStatusModel(log),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.status.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case StatusScreen:
		status, cmd := m.status.Update(msg)
		m.status = status

		return m, cmd
	}

	return m, nil
}

func (m AppModel) View() tea.View {
	if m.status.width == 0 {
		return tea.NewView("")
	}

	const leftWidth = 44
	const rightPadding = 3

	banner := lipgloss.NewStyle().
		PaddingTop(1).
		Render(Banner())

	nowPlaying := lipgloss.NewStyle().
		MarginTop(1).
		PaddingLeft(1).
		Foreground(lipgloss.Color("99")).
		Render("♫  Playing\n" + "   The Astounding Eyes of Rita \n Relaxxx")

	left := lipgloss.NewStyle().
		Width(leftWidth).
		Render(lipgloss.JoinVertical(lipgloss.Left, banner, nowPlaying))

	rightWidth := m.status.width - leftWidth
	if rightWidth < 40 {
		rightWidth = 40
	}

	// contentWidth is what text inside the right panel can actually use
	contentWidth := rightWidth - rightPadding
	if contentWidth < 20 {
		contentWidth = 20
	}

	right := lipgloss.NewStyle().
		Width(rightWidth).
		PaddingLeft(rightPadding).
		PaddingTop(1).
		Render(m.status.rightView(contentWidth))

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		right,
	)

	return tea.NewView(content)
}

func RunApp(log *logger.Logger) error {
	model := NewApp(log)
	program := tea.NewProgram(model)
	_, err := program.Run()

	return err
}