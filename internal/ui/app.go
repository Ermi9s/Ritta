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

	left := lipgloss.NewStyle().
		PaddingTop(1).
		Render(Banner())

	rightWidth := m.status.width - 44

	if rightWidth < 40 {
		rightWidth = 40
	}

	right := lipgloss.NewStyle().
		Width(rightWidth).
		PaddingLeft(3).
		PaddingTop(1).
		Render(m.status.rightView())

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