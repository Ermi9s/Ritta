package ui

import (
	"strings"

	txtin "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)


type InitializeModel struct {
	inputs  []txtin.Model
	cursor  int
	sourceType int
}


func NewInitModel() InitializeModel {
	localDir := txtin.New()
	localDir.Placeholder = "~/projects/project"
	localDir.CharLimit = 200
	localDir.SetWidth(50)

	remoteDir := txtin.New()
	remoteDir.Placeholder = "~/srv/project"
	remoteDir.CharLimit = 200
	remoteDir.SetWidth(50)

	gitRepo := txtin.New()
	gitRepo.Placeholder = "git@github.com:something/project.git"
	gitRepo.CharLimit = 300
	gitRepo.SetWidth(50)

	branch := txtin.New()
	branch.Placeholder = "main"
	branch.CharLimit = 100
	branch.SetWidth(50)


	return InitializeModel{
		inputs: []txtin.Model{
			localDir,  // 0
			remoteDir, // 1
			gitRepo,   // 2
			branch,    // 3
		},
		cursor:     0,
		sourceType: SourceExisting,
	}
}

const (
	SourceExisting = 0
	SourceGit = 1
)

func (m InitializeModel) Init() tea.Cmd {
	return nil
}

func (m InitializeModel) maxCursor() int {
	if m.sourceType == SourceGit {
		return 5
	}

	return 3
}

func (m InitializeModel) inputIndex() int {
	if m.sourceType == SourceExisting {
		switch m.cursor {
		case 1:
			return 0 // local
		case 2:
			return 1 // remote
		}
	}

	if m.sourceType == SourceGit {
		switch m.cursor {
		case 1:
			return 2 // git repo
		case 2:
			return 3 // branch
		case 3:
			return 0 // local
		case 4:
			return 1 // remote
		}
	}

	return -1
}

func (m InitializeModel) rightView() string {
	var b strings.Builder

	b.WriteString(homeTitleStyle.Render("INITIALIZE"))
	b.WriteString("\n\n")

	// Source type
	b.WriteString(homeDimStyle.Render("Source type"))
	b.WriteString("\n\n")

	existing := "  Existing"
	git := "  Git"

	if m.sourceType == SourceExisting {
		existing = "❯ Existing"
	} else {
		git = "❯ Git"
	}

	b.WriteString(existing)
	b.WriteByte('\n')
	b.WriteString(git)
	b.WriteString("\n\n")

	isGit := 0
	if m.sourceType == SourceGit {
		b.WriteString(homeDimStyle.Render("Git repository"))
		b.WriteByte('\n')

		if m.cursor == 1 {
			b.WriteString("❯ ")
		} else {
			b.WriteString("  ")
		}

		b.WriteString(m.inputs[2].View())
		b.WriteString("\n\n")

		b.WriteString(homeDimStyle.Render("Branch"))
		b.WriteByte('\n')

		if m.cursor == 2 {
			b.WriteString("❯ ")
		} else {
			b.WriteString("  ")
		}

		b.WriteString(m.inputs[3].View())
		b.WriteString("\n\n")
		isGit = 2
	}

	// Local directory
	b.WriteString(homeDimStyle.Render("Local directory"))
	b.WriteByte('\n')

	if m.cursor == 1 + isGit {
		b.WriteString("❯ ")
	} else {
		b.WriteString("  ")
	}

	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")



	b.WriteString(homeDimStyle.Render("Remote directory"))
	b.WriteByte('\n')

	if m.cursor == 2 + isGit {
		b.WriteString("❯ ")
	} else {
		b.WriteString("  ")
	}

	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	if m.cursor == 3 + isGit {
		b.WriteString("❯ ")
	} else {
		b.WriteString("  ")
	}

	b.WriteString(homeDimStyle.Render("Continue"))

	b.WriteString("\n\n")
	b.WriteString(
		homeDimStyle.Render(
			"↑/↓ Navigate(bottom)   ←/→ Select(git or existing)   Enter Continue   Esc Back",
		),
	)

	return b.String()
}


func (m *InitializeModel) updateFocus() tea.Cmd {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	switch {
	case m.sourceType == SourceExisting && m.cursor == 1:
		return m.inputs[0].Focus()

	case m.sourceType == SourceExisting && m.cursor == 2:
		return m.inputs[1].Focus()

	case m.sourceType == SourceGit && m.cursor == 1:
		return m.inputs[2].Focus()

	case m.sourceType == SourceGit && m.cursor == 2:
		return m.inputs[3].Focus()

	case m.sourceType == SourceGit && m.cursor == 3:
		return m.inputs[0].Focus()
	case m.sourceType == SourceGit && m.cursor == 4:
		return m.inputs[1].Focus()
	}

	return nil
}

func (m InitializeModel) Update(msg tea.Msg) (InitializeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {

		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			return m, func() tea.Msg {
				return BackToHomeMsg{}
			}

		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, m.updateFocus()

		case "down", "tab":
			if m.cursor < m.maxCursor() {
				m.cursor++
			}
			return m, m.updateFocus()

		case "left":
			if m.cursor == 0 && m.sourceType > SourceExisting {
				m.sourceType--
				return m, m.updateFocus()
			}
			return m, nil

		case "right":
			if m.cursor == 0 && m.sourceType < SourceGit {
				m.sourceType++
				return m, m.updateFocus()
			}
			return m, nil
		case "enter":
			if m.cursor == 0 {
				m.cursor = 1
				return m, m.updateFocus()
			}

			if m.cursor == m.maxCursor() {
				return m, m.submit
			}
		}
	}


	index := m.inputIndex()
	if index >= 0 {
		var cmd tea.Cmd
		m.inputs[index], cmd = m.inputs[index].Update(msg)

		return m, cmd
	}

	return m, nil
}

type InitializeRequest struct {
	SourceType int
	LocalDir   string
	RemoteDir  string
}

func (m InitializeModel) submit() tea.Msg {
	return InitializeRequest{
		SourceType: m.sourceType,
		LocalDir:   m.inputs[0].Value(),
		RemoteDir:  m.inputs[1].Value(),
	}
}