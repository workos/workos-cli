package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/workos/workos-go/v4/pkg/oauthcredentials"
)

var choices = []string{
	string(oauthcredentials.AppleOauth),
	string(oauthcredentials.GithubOauth),
	string(oauthcredentials.GoogleOauth),
	string(oauthcredentials.MicrosoftOauth),
}

type OauthCredentialTypeModel struct {
	cursor int
	Choice string
}

func (m OauthCredentialTypeModel) Init() tea.Cmd {
	return nil
}

func (m OauthCredentialTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "enter":
			m.Choice = choices[m.cursor]
			return m, tea.Quit
		case "down", "j":
			m.cursor++
			if m.cursor >= len(choices) {
				m.cursor = 0
			}
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(choices) - 1
			}
		}
	}

	return m, nil
}

// View implements tea.Model.
func (m OauthCredentialTypeModel) View() string {
	s := strings.Builder{}
	s.WriteString("Which OAuth Authentication Method would you like to create?\n\n")

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")

	}
	s.WriteString("\n(press q to quit)\n")
	return s.String()
}
