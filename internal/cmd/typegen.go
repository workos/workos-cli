package cmd

import (
	"fmt"
	"slices"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/typegen"
)

const (
	FlagLanguage  = "language"
	FlagResources = "resources"
)

// CLI
var typegenCmd = &cobra.Command{
	Use:     "generate-types",
	Short:   "Generate types for WorkOS resources",
	Long:    "A tool to generate type definitions for your WorkOS resources.",
	Example: "workos generate-types --language typescript --resources permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		language, err := cmd.Flags().GetString(FlagLanguage)
		if err != nil {
			return errors.New("Invalid language flag")
		}
		resources, err := cmd.Flags().GetStringArray(FlagResources)
		if err != nil {
			return errors.New("Invalid resources flag")
		}

		var languageGenerator typegen.LanguageTypeGenerator
		switch language {
		case "typescript":
			languageGenerator = typegen.NewTypeScriptTypeGenerator(cmd.OutOrStdout())
		case "python":
			languageGenerator = typegen.NewPythonTypeGenerator(cmd.OutOrStdout())
		default:
			return errors.New(fmt.Sprintf("Invalid language: %s. Valid options are: typescript, python"))
		}

		if err := typegen.GenerateImports(languageGenerator, resources); err != nil {
			return errors.Wrap(err, "error generating imports")
		}

		if slices.Contains(resources, "all") {
			typegen.GeneratePermissionsType(languageGenerator)
			typegen.GenerateRolesType(languageGenerator)
			typegen.GenerateAuditLogsTypes(languageGenerator)
		} else {
			for _, resource := range resources {
				switch resource {
				case "permissions":
					typegen.GeneratePermissionsType(languageGenerator)
				case "roles":
					typegen.GenerateRolesType(languageGenerator)
				case "audit-logs":
					typegen.GenerateAuditLogsTypes(languageGenerator)
				}
			}
		}

		return nil
	},
}

func init() {
	typegenCmd.Flags().StringP(FlagLanguage, "l", "typescript", "Language to to output types in")
	typegenCmd.Flags().StringArrayP(FlagResources, "r", []string{"all"}, "Resources to output types for")
	rootCmd.AddCommand(typegenCmd)
}
