package cmd

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/views"
	"github.com/workos/workos-go/v4/pkg/oauthcredentials"
)

func init() {
	oauthCredCmd.AddCommand(listOauthCreds)
	oauthCredCmd.AddCommand(createOauthCreds)
	rootCmd.AddCommand(oauthCredCmd)
}

var oauthCredCmd = &cobra.Command{
	Use:   "oauthcredentials",
	Short: "Manage OAuth Authentication Methods (get, list, update).",
	Long:  "Get, list, and update oauth authentication methods.",
}

// pass flags instead
var listOauthCreds = &cobra.Command{
	Use:   "list",
	Short: "List oauth credentials",
	Long:  "List oauth credentials",
	Example: `workos oauthcredentials list
workos oauthcredentials list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create the options struct
		oauthCredOpts := oauthcredentials.ListOAuthCredentialsOpts{}

		// List oauth credentials
		oauthCreds, err := oauthcredentials.ListOAuthCredentials(context.Background(), oauthCredOpts)
		if err != nil {
			return fmt.Errorf("error listing oauth credentials: %v", err)
		}

		s := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Render
		t := table.New().Border(lipgloss.NormalBorder()).Width(160).BorderHeader(true)
		t.Headers(s("ID"), s("Type"), s("State"), s("Userland Enabled"))
		trueFalseSymbols := map[bool]string{true: "✅", false: "❌"}

		for _, row := range oauthCreds.Data {
			t.Row(
				row.ID,
				string(row.Type),
				string(row.State),
				trueFalseSymbols[row.IsUserlandEnabled],
			)
		}

		fmt.Println(t.Render())
		return nil
	},
}

var createOauthCreds = &cobra.Command{
	Use:     "create",
	Short:   "Create and oauth authentication method",
	Long:    "Create an oauth authentication method",
	Example: `workos oauthcredentials create`,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(views.OauthCredentialTypeModel{})
		m, err := p.Run()

		if err != nil {
			fmt.Println("Oh no:", err)
			return err
		}

		m, ok := m.(views.OauthCredentialTypeModel)
		choice := m.(views.OauthCredentialTypeModel).Choice

		if !ok || choice == "" {
			return fmt.Errorf("error choosing method type: %v", err)
		}

		oauthCredOpts := oauthcredentials.CreateOAuthCredentialOpts{
			Type: oauthcredentials.OAuthConnectionType(choice),
		}

		oauthCreds, err := oauthcredentials.CreateOAuthCredential(context.Background(), oauthCredOpts)

		if err != nil {
			return fmt.Errorf("error create oauth auth method: %v", err)
		}

		printOauthMethods([]oauthcredentials.OAuthCredential{oauthCreds})
		return nil
	},
}

// pass flags instead
var updateOauthCreds = &cobra.Command{
	Use:   "update",
	Short: "Update oauth credentials",
	Long:  "Update oauth credentials).",
	Example: `workos oauthcredentials update
workos oauthcredentials list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create the options struct
		oauthCredOpts := oauthcredentials.ListOAuthCredentialsOpts{}

		// List oauth credentials
		oauthCreds, err := oauthcredentials.ListOAuthCredentials(context.Background(), oauthCredOpts)
		if err != nil {
			return fmt.Errorf("error listing oauth credentials: %v", err)
		}

		printOauthMethods(oauthCreds.Data)

		return nil
	},
}

func printOauthMethods(methods []oauthcredentials.OAuthCredential) {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00")).Render
	t := table.New().Border(lipgloss.NormalBorder()).Width(160).BorderHeader(true)
	t.Headers(s("ID"), s("Type"), s("State"), s("Userland Enabled"))
	trueFalseSymbols := map[bool]string{true: "✅", false: "❌"}

	for _, row := range methods {
		t.Row(
			row.ID,
			string(row.Type),
			string(row.State),
			trueFalseSymbols[row.IsUserlandEnabled],
		)
	}

	fmt.Println(t.Render())
}
