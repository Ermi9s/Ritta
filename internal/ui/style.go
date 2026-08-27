package ui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	HomeTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#C4B5FD"))

	HomeSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#A78BFA"))

	HomeNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A1A1AA"))

	HomeDescriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D4D4D8")).
				Padding(1, 2)

	HomeDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#71717A"))

	HomeMenuStyle = lipgloss.NewStyle().
			Padding(1, 2)

	HomePanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3F3F46")).
			Padding(1, 2)

	HomeFooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#71717A")).
			PaddingTop(1)

	// Log styles
	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4ADE80"))

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F87171"))

	WarningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FACC15"))

	InfoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#0098c2"))

	CommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#67E8F9"))

	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A1A1AA"))
)

func Success(message string) {
	fmt.Printf("%s %s\n",
		SuccessStyle.Render("✓"),
		SuccessStyle.Render(message),
	)
}

func Error(message string) {
	fmt.Printf("%s %s\n",
		ErrorStyle.Render("✗"),
		ErrorStyle.Render(message),
	)
}

func Warning(message string) {
	fmt.Printf("%s %s\n",
		WarningStyle.Render("!"),
		WarningStyle.Render(message),
	)
}

func Info(message string) {
	fmt.Printf("%s %s\n",
		InfoStyle.Render(":) "),
		InfoStyle.Render(message),
	)
}

func Command(command string) {
	fmt.Printf("    %s\n", CommandStyle.Render(command))
}