package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/workos/workos-cli/internal/list"
	"github.com/workos/workos-cli/internal/views/forms"
	"github.com/workos/workos-cli/internal/views/tables"

	"github.com/spf13/cobra"
	"github.com/workos/workos-go/v4/pkg/organizations"
)

const (
	FlagDomain = "domain"
)

func init() {
	orgCmd.AddCommand(createOrgCmd)
	orgCmd.AddCommand(updateOrgCmd)
	orgCmd.AddCommand(getOrgCmd)
	orgCmd.AddCommand(listOrgCmd)
	orgCmd.AddCommand(deleteOrgCmd)
	rootCmd.AddCommand(orgCmd)
	listOrgCmd.Flags().String(FlagDomain, "", "Filter by domain")
	listOrgCmd.Flags().String(list.FlagAfter, "", "Cursor for results after a specific item")
	listOrgCmd.Flags().String(list.FlagBefore, "", "Cursor for results before a specific item")
	listOrgCmd.Flags().Int(list.FlagLimit, 0, "Limit the number of results")
	listOrgCmd.Flags().String(list.FlagOrder, "", "Order of results (asc or desc)")
}

var orgCmd = &cobra.Command{
	Use:   "organization",
	Short: "Manage organizations (create, update, delete, etc).",
	Long:  "Create, update, and delete organizations and manage organization domain policies.",
}

var createOrgCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new organization",
	Long:    "Create a new organization",
	Example: "workos organization create",
	RunE: func(cmd *cobra.Command, args []string) error {
		orgCreateOpts, err := forms.OrganizationCreate()

		if err != nil {
			return fmt.Errorf("error getting create org data: %v", err)
		}

		org, err := organizations.CreateOrganization(
			context.Background(),
			orgCreateOpts,
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
	Use:     "update <organization_id> <name> [domain] [state]",
	Short:   "Update an organization",
	Long:    "Update an organization's domain or the verification state of its domains (verified or pending).",
	Example: "workos organization update org_01EHZNVPK3SFK441A1RGBFSHRT FooCorp foo-corp.com pending",
	Args:    cobra.RangeArgs(2, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		organizationId := args[0]
		name := args[1]
		var domainData []organizations.OrganizationDomainData

		if len(args) > 2 {
			domainData = append(domainData, organizations.OrganizationDomainData{Domain: args[2], State: "verified"})
		}
		if len(args) > 3 {
			state := organizations.OrganizationDomainDataState(args[3])
			if len(domainData) > 0 {
				domainData[0].State = state
			} else {
				domainData = append(domainData, organizations.OrganizationDomainData{State: state})
			}
		}

		orgOpts := organizations.UpdateOrganizationOpts{
			Organization: organizationId,
			Name:         name,
		}

		if len(domainData) > 0 {
			orgOpts.DomainData = domainData
		}
		org, err := organizations.UpdateOrganization(
			context.Background(),
			orgOpts,
		)

		if err != nil {
			return fmt.Errorf("error updating organization: %v", err)
		}
		fmt.Println(org)
		orgJson, _ := json.MarshalIndent(org, "", "  ")
		fmt.Printf("Updated organization:\n%s\n", string(orgJson))
		return nil
	},
}

var getOrgCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get an organization",
	Long:    "Get an organization by id. Find the organization's id by listing your organizations.",
	Example: `workos organization get <organization_id>`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		organizationId := args[0]
		org, err := organizations.GetOrganization(
			context.Background(),
			organizations.GetOrganizationOpts{
				Organization: organizationId,
			},
		)

		if err != nil {
			return fmt.Errorf("error getting organization: %v", err)
		}
		orgJson, _ := json.MarshalIndent(org, "", "  ")
		fmt.Printf("Organization:\n%s\n", string(orgJson))
		return nil
	},
}

var listOrgCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations with optional filters",
	Long:  "List organizations, optionally filtering by domain, limit, before/after cursor, and order (asc/desc).",
	Example: `workos organization list --domain foo-corp.com --limit 10 --before cursor --order desc
workos organization list --domain foo-corp.com --after cursor --order asc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		after, _ := cmd.Flags().GetString(list.FlagAfter)
		before, _ := cmd.Flags().GetString(list.FlagBefore)
		domain, _ := cmd.Flags().GetString(FlagDomain)
		limit, _ := cmd.Flags().GetInt(list.FlagLimit)
		order, _ := cmd.Flags().GetString(list.FlagOrder)

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

		org, err := organizations.ListOrganizations(
			context.Background(),
			organizations.ListOrganizationsOpts{
				Domains: domains,
				Limit:   limit,
				Before:  before,
				After:   after,
				Order:   orgOrder,
			},
		)
		if err != nil {
			return fmt.Errorf("error listing organizations: %v", err)
		}

		fmt.Println(tables.OrganizationList(org))
		return nil
	},
}

var deleteOrgCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete an organization",
	Long:    "Delete an organization by id. Find the organization's id by listing your organizations.",
	Example: `workos organization delete <organization_id>`,
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
