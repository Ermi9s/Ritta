package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type HomeModel struct {
	cursor int
	width  int
	height int
}

type HomeOption struct {
	Name        string
	Description string
}

func NewHomeModel() HomeModel {
	return HomeModel{
		cursor: 0,
	}
}

func (m HomeModel) Init() tea.Cmd {
	return nil
}


func (m HomeModel) rightView() string {
	selected := homeOptions[m.cursor]

	var menu strings.Builder

	// Header
	menu.WriteString(
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B8A1FF")).
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("#376E88")).
			Padding(0, 1).
			Render("RITTA"),
	)

	menu.WriteString("\n\n")

	// Options
	for i, option := range homeOptions {
		prefix := "  "
		style := homeNormalStyle

		if i == m.cursor {
			prefix = "❯ "
			style = homeSelectedStyle
		}

		menu.WriteString(
			style.Render(prefix + option.Name),
		)

		menu.WriteByte('\n')
	}

	menu.WriteString("\n")

	// Divider
	menu.WriteString(
		homeDimStyle.Render("────────────────────────────────"),
	)

	menu.WriteString("\n\n")

	// Selected option
	menu.WriteString(
		homeTitleStyle.Render(selected.Name),
	)

	menu.WriteString("\n\n")

	// Description
	descriptionWidth := 50

	description := lipgloss.NewStyle().
		Width(descriptionWidth).
		Foreground(lipgloss.Color("#A1A1AA")).
		Render(selected.Description)

	menu.WriteString(description)

	menu.WriteString("\n\n")

	// Shortcuts
	menu.WriteString(
		homeDimStyle.Render(
			"↑/↓ Navigate   Enter Select   Q Quit",
		),
	)

	return menu.String()
}

func (m HomeModel) Update(msg tea.Msg) (HomeModel, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:

		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(homeOptions)-1 {
				m.cursor++
			}

		case "enter":
			return m.handleSelection()
		}
	}

	return m, nil
}

func (m HomeModel) handleSelection() (HomeModel, tea.Cmd) {
	switch m.cursor {

	case 0:
		return m, func() tea.Msg {
			return OpenInitializeMsg{}
		}

	case 1:
		return m, func() tea.Msg {
			return OpenDeployMsg{}
		}

	case 2:
		return m, func() tea.Msg {
			return OpenValidateMsg{}
		}

	case 3:
		return m, func() tea.Msg {
			return OpenConfigureMsg{}
		}

	case 4:
		return m, tea.Quit
	}

	return m, nil
}