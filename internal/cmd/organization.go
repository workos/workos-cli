package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/workos/workos-go/v4/pkg/organizations"
)

func init() {
	orgCmd.AddCommand(createOrgCmd)
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
	Use:     "update <name> <domain> [state]",
	Short:   "Update an organization with a specified name and domain",
	Long:    "Update an organization with a specified name and domain. Optionally, specify the state of the domain (verified or pending).",
	Example: "workos organization update FooCorp foo-corp.com pending",
	Args:    cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		organziation := args[0]
		name := args[1]
		domain := args[2]
		state := organizations.Pending
		if len(args) == 4 {
			state = organizations.OrganizationDomainDataState(args[2])
		}

		org, err := organizations.UpdateOrganization(
			context.Background(),
			organizations.UpdateOrganizationOpts{
				Organization: organziation,
				Name:         name,
				DomainData: []organizations.OrganizationDomainData{
					{
						Domain: domain,
						State:  state,
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("error updating organization: %v", err)
		}

		orgJson, _ := json.MarshalIndent(org, "", "  ")
		fmt.Printf("Updated organization:\n%s\n", string(orgJson))
		return nil
	},
}

var listOrgCmd = &cobra.Command{

	Use:     "list",
	Short:   "List an organizations with a specified domain",
	Long:    "List organizations, can filter by domain, and limit the results and order them in asc and desc.",
	Example: "workos organization list foo-corp.com desc",
	Args:    cobra.MaximumNArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		var domains []string
		var limit int
		var before, after string
		var order organizations.Order

		if len(args) > 0 {
			domains = []string{args[0]}
		}

		if len(args) > 1 {
			var err error
			limit, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid limit value: %v", err)
			}
		}

		if len(args) > 2 {
			switch args[2] {
			case "before":
				before = args[3]
			case "after":
				after = args[3]
			default:
				return fmt.Errorf("third argument must be 'before' or 'after'")
			}
		}

		if len(args) > 3 {
			switch args[3] {
			case "asc":
				order = organizations.Asc
			case "desc":
				order = organizations.Desc
			default:
				return fmt.Errorf("fourth argument must be 'asc' or 'desc'")
			}
		}
		org, err := organizations.ListOrganizations(
			context.Background(),
			organizations.ListOrganizationsOpts{
				Domains: domains,
				Limit:   limit,
				Before:  before,
				After:   after,
				Order:   order,
			},
		)
		if err != nil {
			return fmt.Errorf("error listing organizations: %v", err)
		}

		orgJson, _ := json.MarshalIndent(org, "", "  ")
		fmt.Printf("Organization:\n%s\n", string(orgJson))
		return nil
	},
}
