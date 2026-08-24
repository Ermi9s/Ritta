package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Screen int

const (
	HomeScreen Screen = iota
	InitializeScreen
	DeployScreen
	ValidateScreen
	ConfigureScreen
)

type OpenInitializeMsg struct{}
type OpenDeployMsg struct{}
type OpenValidateMsg struct{}
type OpenConfigureMsg struct{}

type BackToHomeMsg struct{}

type AppModel struct {
	screen     Screen
	home       HomeModel
	initialize InitializeModel
}

func NewApp() AppModel {
	return AppModel{
		screen: HomeScreen,
		home:   NewHomeModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch m.screen {

	case HomeScreen:

		updatedHome, cmd := m.home.Update(msg)
		m.home = updatedHome

		switch msg.(type) {

		case OpenInitializeMsg:
			m.screen = InitializeScreen
			m.initialize = NewInitModel()

			return m, nil

		case OpenDeployMsg:
			m.screen = DeployScreen

			return m, nil

		case OpenValidateMsg:
			m.screen = ValidateScreen

			return m, nil

		case OpenConfigureMsg:
			m.screen = ConfigureScreen

			return m, nil
		}

		return m, cmd

	case InitializeScreen:
		updatedInit, cmd := m.initialize.Update(msg)
		m.initialize = updatedInit

		switch msg := msg.(type) {

		case BackToHomeMsg:
			m.screen = HomeScreen

			return m, nil

		case InitializeRequest:

			// eventually:
			//
			// return m, func() tea.Msg {
			//     err := Initialize(msg)
			//     return InitializeFinishedMsg{Err: err}
			// }

			_ = msg

			return m, func() tea.Msg {
				return BackToHomeMsg{}
			}
		}

		return m, cmd

	case DeployScreen:
		return m, nil

	case ValidateScreen:
		return m, nil

	case ConfigureScreen:
		return m, nil
	}

	return m, nil
}

func (m AppModel) View() tea.View {
	if m.home.width == 0 {
		return tea.NewView("")
	}

	leftTop := Banner()
	left := lipgloss.JoinVertical(
		lipgloss.Left,
		leftTop,
	)

	var right string

	switch m.screen {

	case HomeScreen:
		right = m.home.rightView()

	case InitializeScreen:
		right = m.initialize.rightView()

	case DeployScreen:
		right = "Deploy screen coming soon..."

	case ValidateScreen:
		right = "Validate screen coming soon..."

	case ConfigureScreen:
		right = "Configure screen coming soon..."
	}

	rightStyle := lipgloss.NewStyle().
		Width(54).
		PaddingLeft(3).
		PaddingTop(1)

	right = rightStyle.Render(right)
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		left,
		right,
	)

	return tea.NewView(content)
}

func RunApp() error {
	model := NewApp()

	program := tea.NewProgram(model)

	_, err := program.Run()

	return err
}

func RunHome() error {
	return RunApp()
}