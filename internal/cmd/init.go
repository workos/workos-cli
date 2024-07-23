package cmd

import (
	"errors"
	"github.com/workos/workos-cli/internal/printer"
	"regexp"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/config"
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String(FlagEndpoint, "", "Override the API endpoint")
}

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize the CLI",
	Long:    "Initialize the CLI by configuring an API key for it to use.",
	Example: "workos init",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			name     string
			envType  string
			apiKey   string
			endpoint string
		)

		endpoint, err := cmd.Flags().GetString(FlagEndpoint)
		if err != nil {
			return err
		}

		if len(args) > 0 {
			name = args[0]

			if len(args) > 1 {
				apiKey = args[1]
			} else {
				return errors.New("a valid API key is required")
			}

			if len(args) > 2 {
				endpoint = args[2]
			}
		} else {
			err := huh.NewInput().
				Title("Enter a name for the new environment (e.g. local, staging, etc).").
				Value(&name).
				Validate(func(s string) error {
					if !regexp.MustCompile(EnvironmentNameRegex).Match([]byte(s)) {
						return errors.New("name must only contain lowercase alphanumeric characters and hyphens (-) or underscores (_)")
					}
					return nil
				}).
				Run()
			if err != nil {
				return err
			}

			err = huh.NewSelect[string]().
				Title("Select the type of environment.").
				Options(
					huh.NewOption(EnvironmentTypeProduction, EnvironmentTypeProduction),
					huh.NewOption(EnvironmentTypeSandbox, EnvironmentTypeSandbox),
				).
				Value(&envType).
				Run()
			if err != nil {
				return err
			}

			err = huh.NewInput().
				Title("Enter a valid API key for the environment.").
				Value(&apiKey).
				Run()
			if err != nil {
				return err
			}
		}

		printer.PrintMsg("creating ~/.workos.json")
		envMap := make(map[string]config.Environment)
		envMap[name] = config.Environment{
			ApiKey:   apiKey,
			Name:     name,
			Type:     envType,
			Endpoint: endpoint,
		}
		newConfig := config.Config{
			ActiveEnvironment: name,
			Environments:      envMap,
		}

		err = newConfig.Write()
		if err != nil {
			return err
		}

		printer.PrintMsg("WorkOS CLI initialized")
		return nil
	},
}
