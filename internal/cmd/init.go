package cmd

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/config"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize the CLI for use",
	Long:    "Initialize the CLI for use, including configuring an environment and API key.",
	Example: "workos init",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			apiKey      string
			name        string
			environment string
			endpoint    string
		)

		err := huh.NewInput().
			Title("Enter an API key.").
			Value(&apiKey).
			Run()
		if err != nil {
			return err
		}

		err = huh.NewInput().
			Title("Give this API key a unique name (e.g. john-local-dev).").
			Value(&name).
			Validate(func(s string) error {
				if !regexp.MustCompile("[a-z0-9\\-_]+").Match([]byte(s)) {
					return errors.New("the name can only contain alphanumeric characters and hyphens (-) or underscores (_)")
				}
				return nil
			}).
			Run()
		if err != nil {
			return err
		}

		err = huh.NewInput().
			Title("What environment is this API key for (e.g. Production, Sandbox, etc.)?").
			Value(&environment).
			Run()
		if err != nil {
			return err
		}

		err = huh.NewInput().
			Title("Enter an API endpoint (optional, defaults to https://api.workos.com).").
			Value(&endpoint).
			Run()
		if err != nil {
			return err
		}

		fmt.Println("creating ~/.workos.json")
		apiKeyMap := make(map[string]config.ApiKey)
		apiKeyMap[name] = config.ApiKey{
			Value:       apiKey,
			Name:        name,
			Environment: environment,
			Endpoint:    endpoint,
		}
		newConfig := config.Config{
			ActiveApiKey: name,
			ApiKeys:      apiKeyMap,
		}

		err = newConfig.Write()
		if err != nil {
			return err
		}

		fmt.Println("WorkOS CLI initialized")
		return nil
	},
}
