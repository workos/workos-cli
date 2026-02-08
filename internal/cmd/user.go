package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/list"
	"github.com/workos/workos-cli/internal/printer"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

const (
	flagEmail         = "email"
	flagFirstName     = "first-name"
	flagLastName      = "last-name"
	flagEmailVerified = "email-verified"
	flagPassword      = "password"
	flagExternalID    = "external-id"
	flagOrganization  = "organization"
)

func init() {
	rootCmd.AddCommand(newUserCmd())
}

// newUserCmd returns the user command with all subcommands.
func newUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users (get, list, update, delete, etc).",
		Long:  "Get, list, update, and delete users.",
	}

	cmd.AddCommand(newGetUserCmd())
	cmd.AddCommand(newListUsersCmd())
	cmd.AddCommand(newUpdateUserCmd())
	cmd.AddCommand(newDeleteUserCmd())

	return cmd
}

// newGetUserCmd returns a command to get a user by ID.
func newGetUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <user_id>",
		Short:   "Get a user by ID",
		Long:    "Get a user by their unique identifier.",
		Example: "workos user get user_01E3JC5F5Z1YJNPGVYWV9SX6GH",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID := args[0]

			user, err := usermanagement.GetUser(
				cmd.Context(),
				usermanagement.GetUserOpts{
					User: userID,
				},
			)
			if err != nil {
				return fmt.Errorf("get user: %w", err)
			}

			printer.PrintJson(user)
			return nil
		},
	}
}

// newListUsersCmd returns a command to list users with optional filters.
func newListUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users with optional filters",
		Long:  "List users, optionally filtering by email, organization, limit, before/after cursor, and order (asc/desc).",
		Example: `workos user list --email user@example.com --limit 10
workos user list --organization org_01EHZNVPK3SFK441A1RGBFSHRT --order desc`,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			email, err := flags.GetString(flagEmail)
			if err != nil {
				return fmt.Errorf("email flag: %w", err)
			}

			organizationID, err := flags.GetString(flagOrganization)
			if err != nil {
				return fmt.Errorf("organization flag: %w", err)
			}

			after, err := flags.GetString(list.FlagAfter)
			if err != nil {
				return fmt.Errorf("after flag: %w", err)
			}

			before, err := flags.GetString(list.FlagBefore)
			if err != nil {
				return fmt.Errorf("before flag: %w", err)
			}

			limit, err := flags.GetInt(list.FlagLimit)
			if err != nil {
				return fmt.Errorf("limit flag: %w", err)
			}

			order, err := flags.GetString(list.FlagOrder)
			if err != nil {
				return fmt.Errorf("order flag: %w", err)
			}

			users, err := usermanagement.ListUsers(
				cmd.Context(),
				usermanagement.ListUsersOpts{
					Email:          email,
					OrganizationID: organizationID,
					Limit:          limit,
					Before:         before,
					After:          after,
					Order:          usermanagement.Order(order),
				},
			)
			if err != nil {
				return fmt.Errorf("list users: %w", err)
			}

			tbl := printer.NewTable(140).Headers(
				printer.TableHeader("ID"),
				printer.TableHeader("Email"),
				printer.TableHeader("First Name"),
				printer.TableHeader("Last Name"),
				printer.TableHeader("Verified"),
			)
			for _, user := range users.Data {
				verified := "No"
				if user.EmailVerified {
					verified = "Yes"
				}
				tbl.Row(
					user.ID,
					user.Email,
					user.FirstName,
					user.LastName,
					verified,
				)
			}

			printer.PrintMsg(tbl.Render())
			printer.PrintMsg(fmt.Sprintf("Before: %s", users.ListMetadata.Before))
			printer.PrintMsg(fmt.Sprintf("After: %s", users.ListMetadata.After))
			return nil
		},
	}

	flags := cmd.Flags()
	flags.String(flagEmail, "", "Filter by email")
	flags.String(flagOrganization, "", "Filter by organization ID")
	flags.String(list.FlagAfter, "", "Cursor for results after a specific item")
	flags.String(list.FlagBefore, "", "Cursor for results before a specific item")
	flags.Int(list.FlagLimit, 0, "Limit the number of results")
	flags.String(list.FlagOrder, "", "Order of results (asc or desc)")

	return cmd
}

// newUpdateUserCmd returns a command to update a user's attributes.
func newUpdateUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <user_id>",
		Short: "Update a user",
		Long:  "Update a user's attributes such as first name, last name, email verification status, password, or external ID.",
		Example: `workos user update user_01E3JC5F5Z1YJNPGVYWV9SX6GH --first-name John --last-name Doe
workos user update user_01E3JC5F5Z1YJNPGVYWV9SX6GH --email-verified`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID := args[0]
			flags := cmd.Flags()

			firstName, err := flags.GetString(flagFirstName)
			if err != nil {
				return fmt.Errorf("first-name flag: %w", err)
			}

			lastName, err := flags.GetString(flagLastName)
			if err != nil {
				return fmt.Errorf("last-name flag: %w", err)
			}

			emailVerified, err := flags.GetBool(flagEmailVerified)
			if err != nil {
				return fmt.Errorf("email-verified flag: %w", err)
			}

			password, err := flags.GetString(flagPassword)
			if err != nil {
				return fmt.Errorf("password flag: %w", err)
			}

			externalID, err := flags.GetString(flagExternalID)
			if err != nil {
				return fmt.Errorf("external-id flag: %w", err)
			}

			opts := usermanagement.UpdateUserOpts{
				User:          userID,
				FirstName:     firstName,
				LastName:      lastName,
				EmailVerified: emailVerified,
				Password:      password,
				ExternalID:    externalID,
			}

			user, err := usermanagement.UpdateUser(cmd.Context(), opts)
			if err != nil {
				return fmt.Errorf("update user: %w", err)
			}

			printer.PrintMsg("Updated user")
			printer.PrintJson(user)
			return nil
		},
	}

	flags := cmd.Flags()
	flags.String(flagFirstName, "", "First name")
	flags.String(flagLastName, "", "Last name")
	flags.Bool(flagEmailVerified, false, "Email verification status")
	flags.String(flagPassword, "", "New password")
	flags.String(flagExternalID, "", "External ID")

	return cmd
}

// newDeleteUserCmd returns a command to delete a user by ID.
func newDeleteUserCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <user_id>",
		Short:   "Delete a user",
		Long:    "Delete a user by their unique identifier.",
		Example: "workos user delete user_01E3JC5F5Z1YJNPGVYWV9SX6GH",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			userID := args[0]
			err := usermanagement.DeleteUser(
				cmd.Context(),
				usermanagement.DeleteUserOpts{
					User: userID,
				},
			)
			if err != nil {
				return fmt.Errorf("delete user: %w", err)
			}

			printer.PrintMsg(fmt.Sprintf("Deleted user %s", userID))
			return nil
		},
	}
}
