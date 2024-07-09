package cmd

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/workos/workos-cli/internal/printer"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/config"
)

const (
	EnvironmentNameRegex      = `[a-z0-9\-_]+`
	EnvironmentTypeProduction = "Production"
	EnvironmentTypeSandbox    = "Sandbox"
	FlagEndpoint              = "endpoint"
)

func init() {
	envCmd.AddCommand(addEnvCmd)
	addEnvCmd.Flags().String(FlagEndpoint, "", "Override the API endpoint")
	envCmd.AddCommand(removeEnvCmd)
	envCmd.AddCommand(switchEnvCmd)
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage configured environments",
	Long:  "Add and remove WorkOS environments configured for use with the CLI.",
	Example: `
workos env add
workos env remove
workos env switch`,
	Args: cobra.NoArgs,
}

var addEnvCmd = &cobra.Command{
	Use:     "add [name] [apiKey] [endpoint]",
	Short:   "Configure an environment",
	Long:    "Configure an existing WorkOS environment for use with the CLI.",
	Example: "workos env add",
	Args:    cobra.RangeArgs(0, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfigOrExit()

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

		if len(cfg.Environments) == 0 {
			cfg.Environments = make(map[string]config.Environment)
		}
		cfg.Environments[name] = config.Environment{
			ApiKey:   apiKey,
			Name:     name,
			Type:     envType,
			Endpoint: endpoint,
		}
		err = cfg.Write()
		if err != nil {
			return err
		}

		printer.PrintMsg(fmt.Sprintf("Environment %s added", name))
		return nil
	},
}

var removeEnvCmd = &cobra.Command{
	Use:     "remove [name]",
	Short:   "Remove a configured environment",
	Long:    "Remove a previously configured environment from the WorkOS CLI.",
	Example: "workos env remove",
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfigOrExit()

		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			err := huh.NewInput().
				Title("Enter the name of the environment you would like to remove (e.g. Sandbox)").
				Value(&name).
				Validate(func(s string) error {
					if !regexp.MustCompile(EnvironmentNameRegex).Match([]byte(s)) {
						return errors.New("the name can only contain alphanumeric characters and hyphens (-) or underscores (_)")
					}
					return nil
				}).
				Run()
			if err != nil {
				return err
			}
		}

		if _, ok := config.Environments[name]; !ok {
			return errors.New("the specified environment does not exist")
		}

		delete(config.Environments, name)
		err := config.Write()
		if err != nil {
			return err
		}

		printer.PrintMsg(fmt.Sprintf("Environment %s removed\n", name))
		return nil
	},
}

var switchEnvCmd = &cobra.Command{
	Use:     "switch [name]",
	Short:   "Switch environment",
	Long:    "Switch to using a different environment for subsequent WorkOS CLI commands.",
	Example: "workos env switch",
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfigOrExit()

		var selectedEnvironment string
		environmentOptions := make([]huh.Option[string], len(config.Environments))
		i := 0
		for name, env := range config.Environments {
			label := name
			if env.Type == EnvironmentTypeSandbox {
				label = fmt.Sprintf("%s [%s]", label, EnvironmentTypeSandbox)
			}
			if env.Endpoint != "" {
				label = fmt.Sprintf("%s [%s]", label, env.Endpoint)
			}
			environmentOptions[i] = huh.NewOption(label, name)
			i++
		}

		err := huh.NewSelect[string]().
			Title("Select an environment.").
			Options(environmentOptions...).
			Value(&selectedEnvironment).
			Run()
		if err != nil {
			return err
		}

		config.ActiveEnvironment = selectedEnvironment
		err = config.Write()
		if err != nil {
			return err
		}

		printer.PrintMsg(fmt.Sprintf("Switched to environment %s\n", selectedEnvironment))
		return nil
	},
}
