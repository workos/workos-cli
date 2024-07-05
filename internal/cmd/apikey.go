package cmd

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/config"
)

const (
	ApiKeyRegex           = `[a-z0-9\-_]+`
	EnvironmentProduction = "Production"
	EnvironmentSandbox    = "Sandbox"
)

func init() {
	apiKeyCmd.AddCommand(addApiKeyCmd)
	apiKeyCmd.AddCommand(removeApiKeyCmd)
	apiKeyCmd.AddCommand(switchApiKeyCmd)
	rootCmd.AddCommand(apiKeyCmd)
}

var apiKeyCmd = &cobra.Command{
	Use:     "apikey",
	Short:   "Manage configured API keys",
	Long:    "Add and remove API keys configured for use with the WorkOS CLI.",
	Example: "workos apikey add",
	Args:    cobra.NoArgs,
}

var addApiKeyCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a configured API key",
	Long:    "Configure a new API key for use with the WorkOS CLI.",
	Example: "workos apikey add",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := GetConfigOrExit()

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
				if !regexp.MustCompile(ApiKeyRegex).Match([]byte(s)) {
					return errors.New("the name can only contain alphanumeric characters and hyphens (-) or underscores (_)")
				}
				return nil
			}).
			Run()
		if err != nil {
			return err
		}

		err = huh.NewSelect[string]().
			Title("What type of environment is this API key for?").
			Options(
				huh.NewOption(EnvironmentProduction, EnvironmentProduction),
				huh.NewOption(EnvironmentSandbox, EnvironmentSandbox),
			).
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

		cfg.ApiKeys[name] = config.ApiKey{
			Value:       apiKey,
			Name:        name,
			Environment: environment,
			Endpoint:    endpoint,
		}
		err = cfg.Write()
		if err != nil {
			return err
		}

		fmt.Println("API key added")
		return nil
	},
}

var removeApiKeyCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove a configured API key",
	Long:    "Remove a previously configured API key from use with the WorkOS CLI.",
	Example: "workos apikey remove",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfigOrExit()

		var name string
		err := huh.NewInput().
			Title("Enter the name of the API key you would like to remove (e.g. john-local-dev)").
			Value(&name).
			Validate(func(s string) error {
				if !regexp.MustCompile(ApiKeyRegex).Match([]byte(s)) {
					return errors.New("the name can only contain alphanumeric characters and hyphens (-) or underscores (_)")
				}
				return nil
			}).
			Run()
		if err != nil {
			return err
		}

		if _, ok := config.ApiKeys[name]; !ok {
			return errors.New("the specified API key does not exist")
		}

		delete(config.ApiKeys, name)
		err = config.Write()
		if err != nil {
			return err
		}

		fmt.Println("API key removed")
		return nil
	},
}

var switchApiKeyCmd = &cobra.Command{
	Use:     "switch",
	Short:   "Use the selected API key",
	Long:    "Switch to using the selected API key for subsequent WorkOS CLI commands.",
	Example: "workos apikey switch",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfigOrExit()

		var selectedApiKey string
		apiKeyOptions := make([]huh.Option[string], len(config.ApiKeys))
		i := 0
		for name, apiKey := range config.ApiKeys {
			label := name
			if apiKey.Environment == EnvironmentSandbox {
				label = fmt.Sprintf("%s [%s]", label, EnvironmentSandbox)
			}
			if apiKey.Endpoint != "" {
				label = fmt.Sprintf("%s [%s]", label, apiKey.Endpoint)
			}
			apiKeyOptions[i] = huh.NewOption(label, name)
			i++
		}

		err := huh.NewSelect[string]().
			Title("Select an API key.").
			Options(apiKeyOptions...).
			Value(&selectedApiKey).
			Run()
		if err != nil {
			return err
		}

		config.ActiveApiKey = selectedApiKey
		err = config.Write()
		if err != nil {
			return err
		}

		fmt.Printf("API key %s selected\n", selectedApiKey)
		return nil
	},
}
