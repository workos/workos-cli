package tables

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/workos/workos-go/v4/pkg/organizations"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table         table.Model
	selected      organizations.Organization
	organizations []organizations.Organization
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.selected = m.organizations[m.table.Cursor()]
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func OrganizationListSelector(orgs []organizations.Organization) (organizations.Organization, error) {
	columns := []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Domains", Width: 80},
	}

	rows := []table.Row{}

	for _, org := range orgs {
		var domains []string
		for _, d := range org.Domains {
			domains = append(domains, d.Domain)
		}
		rows = append(rows, table.Row{org.Name, strings.Join(domains, ", ")})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := model{
		table:         t,
		organizations: orgs,
	}

	_, err := tea.NewProgram(m).Run()

	if err != nil {
		return organizations.Organization{}, err
	}

	return m.selected, nil
}
