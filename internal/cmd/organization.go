package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"github.com/workos/workos-go/v4/pkg/organizations"
)

func init() {
	orgCmd.AddCommand(createOrgCmd)
	orgCmd.AddCommand(listOrgCmd)
	rootCmd.AddCommand(orgCmd)
}

var orgCmd = &cobra.Command{
	Use:   "organization",
	Short: "Manage organizations (create, update, delete, etc).",
	Long:  "Create, update, and delete organizations and manage organization domain policies.",
}

// pass flags instead
var listOrgCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations with optional filters",
	Long:  "List organizations, optionally filtering by domain, limit, before/after cursor, and order (asc/desc).",
	Example: `workos organization list --domain foo-corp.com --limit 10 --before cursor --order desc
workos organization list --domain foo-corp.com --after cursor --order asc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieve flag values
		domain, _ := cmd.Flags().GetString("domain")
		limit, _ := cmd.Flags().GetInt("limit")
		before, _ := cmd.Flags().GetString("before")
		after, _ := cmd.Flags().GetString("after")
		order, _ := cmd.Flags().GetString("order")

		var domains []string
		if domain != "" {
			domains = strings.Fields(domain)

		}

		var orgOrder organizations.Order
		switch order {
		case "asc":
			orgOrder = organizations.Asc
		case "desc":
			orgOrder = organizations.Desc
		default:
			if order != "" {
				return fmt.Errorf("invalid order value: must be 'asc' or 'desc'")
			}
		}

		// Create the options struct
		orgOpts := organizations.ListOrganizationsOpts{
			Domains: domains,
			Limit:   limit,
			Before:  before,
			After:   after,
			Order:   orgOrder,
		}

		// List organizations
		org, err := organizations.ListOrganizations(context.Background(), orgOpts)
		if err != nil {
			return fmt.Errorf("error listing organizations: %v", err)
		}

		s := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Render
		t := table.New().Border(lipgloss.NormalBorder()).Width(160).BorderHeader(true)
		t.Headers(s("ID"), s("Name"), s("Domains"), s("Magic Auth"), s("Github Auth"), s("Google Auth"))
		trueFalseSymbols := map[bool]string{true: "✅", false: "❌"}

		for _, row := range org.Data {
			domains := []string{}
			for _, d := range row.Domains {
				domains = append(domains, d.Domain)
			}

			t.Row(
				row.ID,
				row.Name,
				strings.Join(domains, ", "),
				trueFalseSymbols[row.MagicLinkAuthEnabled],
				trueFalseSymbols[row.GithubOauthAuthEnabled],
				trueFalseSymbols[row.GoogleOauthAuthEnabled],
			)
		}

		fmt.Println(t.Render())
		return nil
	},
}

var createOrgCmd = &cobra.Command{
	Use:     "create <name> <domain> [state]",
	Short:   "Create a new organization with a specified name and domain",
	Long:    "Create a new organization with a specified name and domain. Optionally, specify the state of the domain (verified or pending).",
	Example: "workos organization create FooCorp foo-corp.com pending",
	Args:    cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		domain := args[1]
		state := organizations.Pending
		if len(args) == 3 {
			state = organizations.OrganizationDomainDataState(args[2])
		}

		org, err := organizations.CreateOrganization(
			context.Background(),
			organizations.CreateOrganizationOpts{
				Name: name,
				DomainData: []organizations.OrganizationDomainData{
					{
						Domain: domain,
						State:  state,
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("error creating organization: %v", err)
		}

		orgJson, _ := json.MarshalIndent(org, "", "  ")
		fmt.Printf("Created organization:\n%s\n", string(orgJson))
		return nil
	},
}
