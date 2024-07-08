package tables

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/workos/workos-go/v4/pkg/organizations"
)

func OrganizationList(orgs organizations.ListOrganizationsResponse) string {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Render
	t := table.New().Border(lipgloss.NormalBorder()).Width(160).BorderHeader(true)
	t.Headers(s("ID"), s("Name"), s("Domains"))

	for _, row := range orgs.Data {
		var domains []string
		for _, d := range row.Domains {
			domains = append(domains, d.Domain)
		}

		t.Row(
			row.ID,
			row.Name,
			strings.Join(domains, ", "),
		)
	}

	return t.Render()
}
