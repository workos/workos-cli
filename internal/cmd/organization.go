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
		cfg := GetConfigOrExit()
		fmt.Printf("Setting API Key in WorkOS SDK as %s\n", cfg.ApiKeys[cfg.ActiveApiKey].Value)
		organizations.SetAPIKey(cfg.ApiKeys[cfg.ActiveApiKey].Value)

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
