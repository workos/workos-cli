package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/workos/workos-go/v4/pkg/organizations"
)

func init() {
	orgCmd.AddCommand(createOrgCmd)
	orgCmd.AddCommand(updateOrgCmd)
	orgCmd.AddCommand(listOrgCmd)
	orgCmd.AddCommand(deleteOrgCmd)
	rootCmd.AddCommand(orgCmd)
}

var orgCmd = &cobra.Command{
	Use:   "organization",
	Short: "Manage organizations (create, update, delete, etc).",
	Long:  "Create, update, and delete organizations and manage organization domain policies.",
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

var updateOrgCmd = &cobra.Command{
	Use:     "update <organizationId> <name> [domain] [state]",
	Short:   "Update an organization",
	Long:    "Update an organization's domain or the verification state of the domain (verified or pending).",
	Example: "workos organization update FooCorp foo-corp.com pending",
	Args:    cobra.RangeArgs(2, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		organizationId := args[0]
		name := args[1]
		var domain string
		var state organizations.OrganizationDomainDataState
		var hasDomain, hasState bool

		if len(args) > 2 {
			domain = args[2]
			hasDomain = true
		}
		if len(args) > 3 {
			state = organizations.OrganizationDomainDataState(args[3])
			hasState = true
		}

		orgOpts := organizations.UpdateOrganizationOpts{
			Organization: organizationId,
			Name:         name,
		}

		if hasDomain || hasState {
			domainData := organizations.OrganizationDomainData{}
			if hasDomain {
				domainData.Domain = domain
			}
			if hasState {
				domainData.State = state
			}
			orgOpts.DomainData = []organizations.OrganizationDomainData{domainData}
		}
		org, err := organizations.UpdateOrganization(
			context.Background(),
			orgOpts,
		)

		if err != nil {
			return fmt.Errorf("error updating organization: %v", err)
		}

		orgJson, _ := json.MarshalIndent(org, "", "  ")
		fmt.Printf("Updated organization:\n%s\n", string(orgJson))
		return nil
	},
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
			domains = []string{domain}
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

		// Print organizations
		orgJson, _ := json.MarshalIndent(org, "", "  ")
		fmt.Printf("Organizations:\n%s\n", string(orgJson))
		return nil
	},
}

var deleteOrgCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete an organization",
	Long:    "Delete an organization by id. Find the organization's id by listing your organizations.",
	Example: `workos organization delete <organizationId>`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		organizationId := args[0]
		err := organizations.DeleteOrganization(
			context.Background(),
			organizations.DeleteOrganizationOpts{
				Organization: organizationId,
			},
		)

		if err != nil {
			return fmt.Errorf("error deleting organization: %v", err)
		}
		fmt.Printf("Deleted organization %s", organizationId)
		return nil
	},
}

func init() {
	// Define flags
	listOrgCmd.Flags().String("domain", "", "Filter by domain")
	listOrgCmd.Flags().Int("limit", 0, "Limit the number of results")
	listOrgCmd.Flags().String("before", "", "Cursor for results before a specific item")
	listOrgCmd.Flags().String("after", "", "Cursor for results after a specific item")
	listOrgCmd.Flags().String("order", "", "Order of results (asc or desc)")
}
