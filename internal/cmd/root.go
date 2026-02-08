package cmd

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/config"
	"github.com/workos/workos-go/v4/pkg/fga"
	"github.com/workos/workos-go/v4/pkg/organizations"
	"github.com/workos/workos-go/v4/pkg/usermanagement"
)

const (
	// defaultTimeout is the default timeout for commands.
	defaultTimeout = 10 * time.Second
)

var (
	// requestTimeout is the timeout for commands, configurable via a flag.
	requestTimeout time.Duration

	// cmdConfig holds the CLI configuration loaded from disk, including configured environments and the active environment.
	cmdConfig *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "workos",
	Short: "WorkOS Command Line Interface (CLI)",
	Long:  "The WorkOS CLI is a tool to interact with WorkOS APIs via the command line.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if requestTimeout > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
			cobra.OnFinalize(cancel)
			cmd.SetContext(ctx)
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().DurationVar(&requestTimeout, "timeout", defaultTimeout, "Timeout for commands")
}

func SetVersion(version string) {
	rootCmd.Version = version
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func GetConfigOrExit() *config.Config {
	if cmdConfig.ActiveEnvironment == "" {
		log.Fatal("no active environment configured. Run 'workos init'")
	}
	if len(cmdConfig.Environments) == 0 {
		log.Fatal("no environments configured. Run 'workos init'")
	}
	if _, ok := cmdConfig.Environments[cmdConfig.ActiveEnvironment]; !ok {
		log.Fatal("configured active environment is invalid. Run 'workos init'")
	}
	return cmdConfig
}

func initConfig() {
	cmdConfig = config.LoadConfig()
	organizations.SetAPIKey(cmdConfig.Environments[cmdConfig.ActiveEnvironment].ApiKey)
	fga.SetAPIKey(cmdConfig.Environments[cmdConfig.ActiveEnvironment].ApiKey)
	usermanagement.SetAPIKey(cmdConfig.Environments[cmdConfig.ActiveEnvironment].ApiKey)
	if cmdConfig.Environments[cmdConfig.ActiveEnvironment].Endpoint != "" {
		organizations.DefaultClient.Endpoint = cmdConfig.Environments[cmdConfig.ActiveEnvironment].Endpoint
		fga.DefaultClient.Endpoint = cmdConfig.Environments[cmdConfig.ActiveEnvironment].Endpoint
		usermanagement.DefaultClient.Endpoint = cmdConfig.Environments[cmdConfig.ActiveEnvironment].Endpoint
	}
}
