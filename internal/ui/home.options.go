package ui

import "charm.land/lipgloss/v2"

var homeOptions = []HomeOption{
	{
		Name:        "Initialize",
		Description: "Create a Ritta configuration for your project.",
	},
	{
		Name:        "Deploy",
		Description: "Deploy your project to a remote server.",
	},
	{
		Name:        "Validate",
		Description: "Validate your Ritta configuration.",
	},
	{
		Name:        "Configure",
		Description: "Configure your Ritta deployment.",
	},
	{
		Name:        "Exit",
		Description: "Exit Ritta.",
	},
}

var (

	homeTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#C4B5FD"))

	homeSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#A78BFA"))

	homeNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A1A1AA"))

	homeDescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D4D4D8")).
				Padding(1, 2)

	homeDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#71717A"))

	homeMenuStyle = lipgloss.NewStyle().
			Padding(1, 2)

	homePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3F3F46")).
			Padding(1, 2)

	homeFooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#71717A")).
			PaddingTop(1)
)

