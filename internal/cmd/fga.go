package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss/list"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/printer"
	"github.com/workos/workos-go/v4/pkg/fga"
	"github.com/workos/workos-go/v4/pkg/workos_errors"
	"gopkg.in/yaml.v3"
)

const (
	ResourceSeparator = ":"
	RelationSeparator = "#"

	TestFileExtensionJSON = ".test.json"
	TestFileExtensionYAML = ".test.yaml"
	TestFileExtensionYML  = ".test.yml"
)

var resourceTypesFile string
var AcceptedTestFileExtensions = []string{TestFileExtensionJSON, TestFileExtensionYAML, TestFileExtensionYML}
var AcceptedTestFileExtensionErrorMessage = fmt.Sprintf("Accepted test file extensions are: %s", strings.Join(AcceptedTestFileExtensions, ", "))

func init() {
	// resource-types
	listResourceTypesCmd.Flags().Int("limit", 10, "limit the number of results returned")
	listResourceTypesCmd.Flags().String("before", "", "cursor indicating results that occur before a specific result")
	listResourceTypesCmd.Flags().String("after", "", "cursor indicating results that occur after a specific result")
	listResourceTypesCmd.Flags().String("order", "", "order in which a list of results should be returned (asc or desc)")
	resourceTypeCmd.AddCommand(listResourceTypesCmd)
	applyResourceTypesCmd.Flags().StringVarP(&resourceTypesFile, "file", "f", "", "file containing resource type definitions")
	resourceTypeCmd.AddCommand(applyResourceTypesCmd)
	fgaCmd.AddCommand(resourceTypeCmd)

	// warrants
	createWarrantCmd.Flags().StringP("policy", "p", "", "boolean expression to be evaluated for a warrant at the time of a check")
	warrantCmd.AddCommand(createWarrantCmd)
	warrantCmd.AddCommand(deleteWarrantCmd)
	fgaCmd.AddCommand(warrantCmd)

	// check
	checkRelationCmd.Flags().StringP("warrantToken", "w", "", "warrant token to use for check")
	checkRelationCmd.Flags().String("assert", "", "assert that the check is true or false")
	checkRelationCmd.Flags().BoolP("debug", "d", false, "run check in debug mode")
	fgaCmd.AddCommand(checkRelationCmd)

	// resources
	resourceCmd.AddCommand(createResourceCmd)
	listResourcesCmd.Flags().String("type", "", "resource type to filter results by")
	listResourcesCmd.Flags().String("search", "", "search term to filter a list of results by")
	listResourcesCmd.Flags().Int("limit", 10, "limit the number of results returned")
	listResourcesCmd.Flags().String("before", "", "cursor indicating results that occur before a specific result")
	listResourcesCmd.Flags().String("after", "", "cursor indicating results that occur after a specific result")
	listResourcesCmd.Flags().String("order", "", "order in which a list of results should be returned (asc or desc)")
	resourceCmd.AddCommand(listResourcesCmd)
	resourceCmd.AddCommand(updateResourceCmd)
	resourceCmd.AddCommand(deleteResourceCmd)
	fgaCmd.AddCommand(resourceCmd)

	// query
	queryCmd.Flags().StringP("warrantToken", "w", "", "warrant token to use for query")
	queryCmd.Flags().Int("limit", 10, "limit the number of results returned")
	queryCmd.Flags().String("before", "", "cursor indicating results that occur before a specific result")
	queryCmd.Flags().String("after", "", "cursor indicating results that occur after a specific result")
	queryCmd.Flags().String("order", "", "order in which a list of results should be returned (asc or desc)")
	fgaCmd.AddCommand(queryCmd)

	// schema
	convertSchemaCMD.Flags().String("to", "json", "output to (schema or json)")
	convertSchemaCMD.Flags().String("output", "pretty", "output pretty or raw. use raw for machine-readable output or writing to a file")
	schemaCmd.AddCommand(convertSchemaCMD)
	schemaCmd.AddCommand(getSchemaCmd)
	schemaCmd.AddCommand(validateSchemaCmd)
	applySchemaCmd.Flags().BoolP("verbose", "v", false, "print extra details about the request")
	applySchemaCmd.Flags().Bool("strict", false, "fail if there are warnings")
	schemaCmd.AddCommand(applySchemaCmd)
	fgaCmd.AddCommand(schemaCmd)

	testCmd.Flags().Bool("cleanup", false, "Optional: Delete all resources referenced in setup warrants after test run")
	fgaCmd.AddCommand(testCmd)

	rootCmd.AddCommand(fgaCmd)
}

var fgaCmd = &cobra.Command{
	Use:   "fga",
	Short: "Manage FGA resources (resource types, warrants, and resources).",
	Long:  "Manage FGA-specific resources like resource types, warrants, and resources and perform check and query operations to validate your FGA model.",
}

var resourceTypeCmd = &cobra.Command{
	Use:   "resourcetype",
	Short: "Manage your resource types",
	Long:  "List and apply resource types. Resource types are used to define the types of resources in your application and the relations between them.",
}

var listResourceTypesCmd = &cobra.Command{
	Use:     "list",
	Short:   "List resource types",
	Long:    "List resource types, optionally providing common flags to filter and paginate the results.",
	Example: "workos fga resourcetype list --limit=5",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceTypes, err := fga.ListResourceTypes(context.Background(), fga.ListResourceTypesOpts{
			Limit: 100,
		})
		if err != nil {
			return errors.Errorf("error listing resource types: %v", err)
		}

		tbl := printer.NewTable(80).Headers(
			printer.TableHeader("Resource Type"),
		)
		for _, resourceType := range resourceTypes.Data {
			tbl.Row(
				resourceType.Type,
			)
		}

		printer.PrintMsg(tbl.Render())
		printer.PrintMsg(fmt.Sprintf("Before: %s", resourceTypes.ListMetadata.Before))
		printer.PrintMsg(fmt.Sprintf("After: %s", resourceTypes.ListMetadata.After))
		return nil
	},
}

var applyResourceTypesCmd = &cobra.Command{
	Use:     "apply",
	Short:   "Apply a set of resource types",
	Long:    "Apply a set of resource types from a specified file. This command will create any resource types present in the file and delete any resource types that are not.",
	Example: "workos fga resourcetype apply -f resource-types.json",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var bytes []byte
		var err error
		if resourceTypesFile != "" {
			jsonFile, err := os.Open(resourceTypesFile)
			if err != nil {
				return err
			}
			defer jsonFile.Close()

			bytes, err = io.ReadAll(jsonFile)
			if err != nil {
				return err
			}
		} else {
			bytes, err = io.ReadAll(bufio.NewReader(os.Stdin))
			if err != nil {
				return err
			}
		}

		var resourceTypes []fga.UpdateResourceTypeOpts
		err = json.Unmarshal(bytes, &resourceTypes)
		if err != nil {
			return err
		}

		_, err = fga.BatchUpdateResourceTypes(context.Background(), resourceTypes)
		if err != nil {
			return err
		}

		printer.PrintMsg("Resource types updated")
		return nil
	},
}

var warrantCmd = &cobra.Command{
	Use:   "warrant",
	Short: "Manage your warrants",
	Long:  "Create and delete warrants that define the relationships between resources in your application.",
}

var createWarrantCmd = &cobra.Command{
	Use:     "create <subject> <relation> <resource> [policy]",
	Short:   "Create a warrant",
	Long:    "Create a warrant assigning a relation between a subject and a resource, optionally specifying a policy that dictates when the relation applies.",
	Example: "workos fga warrant create user:john owner document:xyz --policy \"region == 'eu'\"",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectType, subjectIdRelation, valid := strings.Cut(args[0], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid subject: %s", args[0])
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, RelationSeparator)
		relation := args[1]
		resourceType, resourceId, valid := strings.Cut(args[2], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid resource: %s", args[0])
		}

		policy, err := cmd.Flags().GetString("policy")
		if err != nil {
			return errors.Errorf("invalid policy flag")
		}

		res, err := fga.WriteWarrant(
			context.Background(),
			fga.WriteWarrantOpts{
				Op:           fga.WarrantOpCreate,
				ResourceType: resourceType,
				ResourceId:   resourceId,
				Relation:     relation,
				Subject: fga.Subject{
					ResourceType: subjectType,
					ResourceId:   subjectId,
					Relation:     subjectRelation,
				},
				Policy: policy,
			},
		)
		if err != nil {
			return errors.Errorf("error creating warrant: %v", err)
		}

		if policy != "" {
			printer.PrintMsg(fmt.Sprintf("Assigned %s %s %s [%s]", args[0], args[1], args[2], policy))
		} else {
			printer.PrintMsg(fmt.Sprintf("Assigned %s %s %s", args[0], args[1], args[2]))
		}
		printer.PrintMsg(fmt.Sprintf("Warrant-Token: %s", res.WarrantToken))
		return nil
	},
}

var deleteWarrantCmd = &cobra.Command{
	Use:     "delete <subject> <relation> <resource>",
	Short:   "Delete a warrant",
	Long:    "Delete a warrant that assigns a relation between a subject and a resource.",
	Example: "workos fga warrant delete user:john owner document:xyz",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectType, subjectIdRelation, valid := strings.Cut(args[0], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid subject: %s", args[0])
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, RelationSeparator)
		relation := args[1]
		resourceType, resourceId, valid := strings.Cut(args[2], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid resource: %s", args[0])
		}

		res, err := fga.WriteWarrant(
			context.Background(),
			fga.WriteWarrantOpts{
				Op:           fga.WarrantOpDelete,
				ResourceType: resourceType,
				ResourceId:   resourceId,
				Relation:     relation,
				Subject: fga.Subject{
					ResourceType: subjectType,
					ResourceId:   subjectId,
					Relation:     subjectRelation,
				},
			},
		)
		if err != nil {
			return errors.Errorf("error removing relation: %v", err)
		}

		printer.PrintMsg(fmt.Sprintf("Removed %s %s %s", args[0], args[1], args[2]))
		printer.PrintMsg(fmt.Sprintf("Warrant-Token: %s", res.WarrantToken))
		return nil
	},
}

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage your resources",
	Long:  "Create, update, read, list and delete resources.",
}

var createResourceCmd = &cobra.Command{
	Use:     "create <resource> [meta]",
	Short:   "Create a new resource",
	Long:    "Create a new resource of a given type, optionally providing an identifier for the resource and/or any metadata to attach to the resource.",
	Example: `workos fga resource create user:john '{"email":"john.doe@workos.com"}'`,
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceType, resourceId, valid := strings.Cut(args[0], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid resource: %s", args[0])
		}

		var meta map[string]interface{}
		var err error
		if len(args) == 2 {
			err = json.Unmarshal([]byte(args[1]), &meta)
			if err != nil {
				return errors.Errorf("invalid resource meta: %s", args[1])
			}
		}

		createdResource, err := fga.CreateResource(context.Background(), fga.CreateResourceOpts{
			ResourceType: resourceType,
			ResourceId:   resourceId,
			Meta:         meta,
		})
		if err != nil {
			return errors.Errorf("error creating resource: %v", err)
		}

		if len(createdResource.Meta) > 0 {
			printer.PrintMsg(fmt.Sprintf("Created resource %s:%s (%v)", createdResource.ResourceType, createdResource.ResourceId, createdResource.Meta))
		} else {
			printer.PrintMsg(fmt.Sprintf("Created resource %s:%s", createdResource.ResourceType, createdResource.ResourceId))
		}

		return nil
	},
}

var listResourcesCmd = &cobra.Command{
	Use:     "list",
	Short:   "List resources",
	Long:    "List resources, optionally specifying the '--type' flag to filter to resources of a specific type or providing common flags to filter and paginate the results.",
	Example: "workos fga resource list --type=user --limit=15",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceType, err := cmd.Flags().GetString("type")
		if err != nil {
			return errors.Errorf("invalid type flag")
		}
		search, err := cmd.Flags().GetString("search")
		if err != nil {
			return errors.Errorf("invalid search flag")
		}
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return errors.Errorf("invalid limit flag")
		}
		before, err := cmd.Flags().GetString("before")
		if err != nil {
			return errors.Errorf("invalid before flag")
		}
		after, err := cmd.Flags().GetString("after")
		if err != nil {
			return errors.Errorf("invalid after flag")
		}
		order, err := cmd.Flags().GetString("order")
		if err != nil {
			return errors.Errorf("invalid order flag")
		}
		var orderFilter fga.Order
		if order != "" {
			if strings.ToLower(order) == "asc" {
				orderFilter = fga.Asc
			} else {
				orderFilter = fga.Desc
			}
		}

		resources, err := fga.ListResources(context.Background(), fga.ListResourcesOpts{
			ResourceType: resourceType,
			Search:       search,
			Limit:        limit,
			Before:       before,
			After:        after,
			Order:        orderFilter,
		})
		if err != nil {
			return errors.Errorf("error listing resources: %v", err)
		}

		tbl := printer.NewTable(120).Headers(
			printer.TableHeader("Resource Type"),
			printer.TableHeader("Resource ID"),
			printer.TableHeader("Meta"),
		)
		for _, resource := range resources.Data {
			metaString, err := json.MarshalIndent(resource.Meta, "", "    ")
			if err != nil {
				return errors.Errorf("error listing resources: %v", err)
			}
			tbl.Row(
				resource.ResourceType,
				resource.ResourceId,
				string(metaString),
			)
		}

		printer.PrintMsg(tbl.Render())
		printer.PrintMsg(fmt.Sprintf("Before: %s", resources.ListMetadata.Before))
		printer.PrintMsg(fmt.Sprintf("After: %s", resources.ListMetadata.After))
		return nil
	},
}

var updateResourceCmd = &cobra.Command{
	Use:     "update <resource> <meta>",
	Short:   "Update a resource",
	Long:    "Update a resource, providing metadata to attach to it.",
	Example: `workos fga resource update user:john '{"email":"john.doe@workos.com"}'`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceType, resourceId, valid := strings.Cut(args[0], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid resource: %s", args[0])
		}

		var meta map[string]interface{}
		err := json.Unmarshal([]byte(args[1]), &meta)
		if err != nil {
			return errors.Errorf("invalid meta: %s", args[1])
		}

		updatedResource, err := fga.UpdateResource(context.Background(), fga.UpdateResourceOpts{
			ResourceType: resourceType,
			ResourceId:   resourceId,
			Meta:         meta,
		})
		if err != nil {
			return errors.Errorf("error updating resource: %v", err)
		}

		printer.PrintMsg(fmt.Sprintf("Updated resource %s:%s", updatedResource.ResourceType, updatedResource.ResourceId))
		return nil
	},
}

var deleteResourceCmd = &cobra.Command{
	Use:     "delete <resource>",
	Short:   "Delete a resource",
	Long:    "Delete a given resource. This will delete any warrants associated with the resource.",
	Example: `workos fga resource delete user:john`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceType, resourceId, valid := strings.Cut(args[0], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid resource: %s", args[0])
		}

		err := fga.DeleteResource(context.Background(), fga.DeleteResourceOpts{
			ResourceType: resourceType,
			ResourceId:   resourceId,
		})
		if err != nil {
			return errors.Errorf("error deleting resource: %v", err)
		}

		printer.PrintMsg(fmt.Sprintf("Deleted resource %s", args[0]))
		return nil
	},
}

var checkRelationCmd = &cobra.Command{
	Use:     "check <subject> <relation> <resource> [context]",
	Short:   "Check for a relation",
	Long:    "Check if a given subject has the specified relation on a given resource, optionally specifying context to use while evaluating the check.",
	Example: `workos fga check user:john owner document:xyz '{"organization": "acme"}'`,
	Args:    cobra.RangeArgs(3, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectType, subjectIdRelation, valid := strings.Cut(args[0], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid subject: %s", args[0])
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, RelationSeparator)
		relation := args[1]
		resourceType, resourceId, valid := strings.Cut(args[2], ResourceSeparator)
		if !valid {
			return errors.Errorf("invalid resource: %s", args[0])
		}

		var policyContext map[string]interface{}
		if len(args) > 3 {
			err := json.Unmarshal([]byte(args[3]), &policyContext)
			if err != nil {
				return errors.Errorf("invalid context: %s", args[3])
			}
		}

		warrantToken, err := cmd.Flags().GetString("warrantToken")
		if err != nil {
			return errors.Wrap(err, "invalid warrantToken flag")
		}
		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return errors.Wrap(err, "invalid debug flag")
		}

		warrantCheck := fga.WarrantCheck{
			ResourceType: resourceType,
			ResourceId:   resourceId,
			Relation:     relation,
			Subject: fga.Subject{
				ResourceType: subjectType,
				ResourceId:   subjectId,
				Relation:     subjectRelation,
			},
			Context: policyContext,
		}
		result, err := fga.Check(
			context.Background(),
			fga.CheckOpts{
				Checks: []fga.WarrantCheck{
					warrantCheck,
				},
				WarrantToken: warrantToken,
				Debug:        debug,
			},
		)
		if err != nil {
			return errors.Errorf("error evaluating check: %v", err)
		}

		warrantCheckString, err := warrantCheckAsString(warrantCheck)
		if err != nil {
			return errors.Errorf("invalid check: %v", err)
		}

		assert, err := cmd.Flags().GetString("assert")
		if err != nil {
			return errors.Wrap(err, "invalid assert flag")
		}
		if assert != "" {
			assertBool, err := strconv.ParseBool(assert)
			if err != nil {
				return errors.Errorf("invalid assertion: %s", assert)
			}

			if assertBool == result.Authorized() {
				printer.PrintMsg(fmt.Sprintf("%s %s", printer.GreenText(printer.Checkmark, fmt.Sprintf("assert %t", assertBool)), warrantCheckString))
			} else {
				printer.PrintMsg(fmt.Sprintf("%s %s", printer.RedText(printer.Cross, fmt.Sprintf("assert %t", assertBool)), warrantCheckString))
				os.Exit(1)
			}
		} else if result.Authorized() {
			printer.PrintMsg(fmt.Sprintf("%s %s", printer.GreenText(printer.Checkmark, fga.CheckResultAuthorized), warrantCheckString))
		} else {
			printer.PrintMsg(fmt.Sprintf("%s %s", printer.RedText(printer.Cross, fga.CheckResultNotAuthorized), warrantCheckString))
		}

		if debug {
			printer.PrintMsg(fmt.Sprintf("Response Time: %dms", result.DebugInfo.ProcessingTime/1000000)) // convert ns to ms
			if result.DebugInfo.DecisionTree != nil {
				printer.PrintMsg(buildDecisionTreeList(*result.DebugInfo.DecisionTree).String())
			}
		}

		return nil
	},
}

var queryCmd = &cobra.Command{
	Use:     "query <query> [context]",
	Short:   "Query for access rules",
	Long:    "Run a query to see which resources a subject has access to or which subjects have access to a resource, optionally specifying context to use while evaluating the query.",
	Example: "workos fga query select document where user:john is owner",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		after, err := cmd.Flags().GetString("after")
		if err != nil {
			return errors.Wrap(err, "invalid after flag")
		}
		before, err := cmd.Flags().GetString("before")
		if err != nil {
			return errors.Wrap(err, "invalid before flag")
		}
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return errors.Wrap(err, "invalid limit flag")
		}
		order, err := cmd.Flags().GetString("order")
		if err != nil {
			return errors.Wrap(err, "invalid order flag")
		}
		warrantToken, err := cmd.Flags().GetString("warrantToken")
		if err != nil {
			return errors.Wrap(err, "invalid warrantToken flag")
		}
		var policyContext map[string]interface{}
		if len(args) > 1 {
			err := json.Unmarshal([]byte(args[1]), &policyContext)
			if err != nil {
				return errors.Errorf("invalid context: %s", args[1])
			}
		}

		result, err := fga.Query(context.Background(), fga.QueryOpts{
			Query:        args[0],
			Context:      policyContext,
			Limit:        limit,
			Before:       before,
			After:        after,
			Order:        fga.Order(order),
			WarrantToken: warrantToken,
		})
		if err != nil {
			return errors.Errorf("error performing query: %v", err)
		}

		tbl := printer.NewTable(120).Headers(
			printer.TableHeader("Resource Type"),
			printer.TableHeader("Resource ID"),
			printer.TableHeader("Relation"),
			printer.TableHeader("Implicit"),
			printer.TableHeader("Meta"),
		)
		for _, queryResult := range result.Data {
			metaString, err := json.MarshalIndent(queryResult.Meta, "", "    ")
			if err != nil {
				return errors.Errorf("error listing resources: %v", err)
			}
			tbl.Row(
				queryResult.ResourceType,
				queryResult.ResourceId,
				queryResult.Relation,
				strconv.FormatBool(queryResult.IsImplicit),
				string(metaString),
			)
		}

		printer.PrintMsg(tbl.Render())
		printer.PrintMsg(fmt.Sprintf("Before: %s", result.ListMetadata.Before))
		printer.PrintMsg(fmt.Sprintf("After: %s", result.ListMetadata.After))
		return nil
	},
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage your schema",
	Long:  "A schema is a set of resource types and relations that define the structure of your application. Use this command to manage your schema.",
}

var convertSchemaCMD = &cobra.Command{
	Use:     "convert <input_file>",
	Short:   "Convert a schema to a JSON representation (or JSON to schema)",
	Long:    "Convert a schema to a JSON representation (or JSON to schema) that can be used to apply resource types.",
	Example: `workos fga schema convert schema.txt --to json`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		to, err := cmd.Flags().GetString("to")
		if err != nil {
			return errors.Wrap(err, "invalid to flag")
		}
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return errors.Wrap(err, "invalid output flag")
		}

		bytes, err := os.ReadFile(args[0])
		if err != nil {
			return errors.Errorf("error reading input file: %v", err)
		}

		var response fga.ConvertSchemaResponse
		switch to {
		case "json":
			schemaString := string(bytes)
			response, err = fga.ConvertSchemaToResourceTypes(context.Background(), fga.ConvertSchemaToResourceTypesOpts{
				Schema: schemaString,
			})
			if err != nil {
				return convertSchemaError(err)
			}
		case "schema":
			var resourceTypesWithVersion fga.ConvertResourceTypesToSchemaOpts
			err = json.Unmarshal(bytes, &resourceTypesWithVersion)
			if err != nil {
				return errors.Errorf("error unmarshalling resource types: %v", err)
			}
			response, err = fga.ConvertResourceTypesToSchema(context.Background(), resourceTypesWithVersion)
			if err != nil {
				return convertSchemaError(err)
			}
		default:
			return errors.Errorf("invalid conversion: %s", to)
		}

		switch output {
		case "pretty":
			printer.PrintMsg("Version:")
			printer.PrintMsg(fmt.Sprintf("%s\n", response.Version))

			if len(response.Warnings) > 0 {
				printer.PrintMsg(printer.YellowText("Warnings:"))
				for _, warning := range response.Warnings {
					printer.PrintMsg(printer.YellowText(warning.Message))
				}
				printer.PrintMsg("\n")
			}

			if response.Schema != nil {
				printer.PrintMsg("Schema:")
				printer.PrintMsg(*response.Schema)
			}

			if response.ResourceTypes != nil {
				printer.PrintMsg("Resource Types:")
				printer.PrintJson(response.ResourceTypes)
			}
			if response.Policies != nil {
				printer.PrintMsg("Policies:")
				for key, policy := range response.Policies {
					printer.PrintMsg(fmt.Sprintf("%s: %s", key, policy))
				}
			}
		case "raw":
			if response.Schema != nil {
				printer.PrintMsg(*response.Schema)
			} else {
				printer.PrintJson(response)
			}
		default:
			return errors.Errorf("invalid output: %s", output)
		}
		return nil
	},
}

var getSchemaCmd = &cobra.Command{
	Use:     "get",
	Short:   "Get the current schema",
	Long:    "Get the current schema, which includes the version, resource types, and policies.",
	Example: `workos fga schema get`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		response, err := fga.GetSchema(context.Background())
		if err != nil {
			return errors.Errorf("error getting schema: %v", err)
		}
		convertResponse, err := fga.ConvertResourceTypesToSchema(context.Background(), fga.ConvertResourceTypesToSchemaOpts(response))
		if err != nil {
			return convertSchemaError(err)
		}

		if convertResponse.Schema != nil {
			printer.PrintMsg(*convertResponse.Schema)
		} else {
			return errors.New("error getting schema: no schema found")
		}
		return nil
	},
}

var validateSchemaCmd = &cobra.Command{
	Use:     "validate <input_file>",
	Short:   "Validate a schema file for issues",
	Long:    "Validate a schema file, printing any warnings or errors found. Only schema files are accepted (not resource type JSON).",
	Example: `workos fga schema validate schema.txt`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bytes, err := os.ReadFile(args[0])
		if err != nil {
			return errors.Errorf("error reading input file: %v", err)
		}

		response, err := fga.ConvertSchemaToResourceTypes(context.Background(), fga.ConvertSchemaToResourceTypesOpts{
			Schema: string(bytes),
		})
		if err != nil {
			return convertSchemaError(err)
		}

		if len(response.Warnings) > 0 {
			printer.PrintMsg(printer.YellowText("Warnings:"))
			for _, warning := range response.Warnings {
				printer.PrintMsg(printer.YellowText("* " + warning.Message))
			}
			return nil
		}

		printer.PrintMsg(printer.GreenText(printer.Checkmark + " No issues"))
		return nil
	},
}

var applySchemaCmd = &cobra.Command{
	Use:     "apply <input_file>",
	Short:   "Apply a schema",
	Long:    "Apply a schema to create or update resource types.",
	Example: `workos fga schema apply schema.txt`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return errors.Wrap(err, "invalid verbose flag")
		}
		strict, err := cmd.Flags().GetBool("strict")
		if err != nil {
			return errors.Wrap(err, "invalid strict flag")
		}

		bytes, err := os.ReadFile(args[0])
		if err != nil {
			return errors.Errorf("error reading input file: %v", err)
		}
		// Convert schema to resource types
		response, err := fga.ConvertSchemaToResourceTypes(context.Background(), fga.ConvertSchemaToResourceTypesOpts{
			Schema: string(bytes),
		})
		if err != nil {
			return convertSchemaError(err)
		}

		if len(response.Warnings) > 0 {
			printer.PrintMsg(printer.YellowText("Warnings:"))
			for _, warning := range response.Warnings {
				printer.PrintMsg(printer.YellowText("* " + warning.Message))
			}
			printer.PrintMsg("\n")
			if strict {
				return errors.New("error applying schema: warnings found (omit --strict to ignore)")
			}
		}

		printer.PrintMsg("applying schema...")

		if verbose {
			printer.PrintJson(response.ResourceTypes)
		}

		resourceTypeOps := make([]fga.UpdateResourceTypeOpts, 0)
		for _, rt := range response.ResourceTypes {
			resourceTypeOps = append(resourceTypeOps, fga.UpdateResourceTypeOpts(rt))
		}

		policyOps := make(map[string]fga.UpdatePolicyOpts, len(response.Policies))
		for key, p := range response.Policies {
			policyOps[key] = fga.UpdatePolicyOpts(p)
		}

		_, err = fga.UpdateSchema(context.Background(), fga.UpdateSchemaOpts{ResourceTypes: resourceTypeOps, Policies: policyOps})
		if err != nil {
			return errors.Errorf("error applying schema: %v", err)
		}

		printer.PrintMsg("Schema applied")
		return nil
	},
}

// Test spec types

type WarrantSetup struct {
	Subject  string `yaml:"subject" json:"subject"`
	Relation string `yaml:"relation" json:"relation"`
	Resource string `yaml:"resource" json:"resource"`
	Policy   string `yaml:"policy,omitempty" json:"policy,omitempty"`
}

type FGATestSetupSection struct {
	Warrants []WarrantSetup `yaml:"warrants" json:"warrants"`
}

type FGATestCheck struct {
	Subject  string                 `yaml:"subject" json:"subject"`
	Relation string                 `yaml:"relation" json:"relation"`
	Resource string                 `yaml:"resource" json:"resource"`
	Context  map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"`
}

type FGATestCase struct {
	Name   string       `yaml:"name" json:"name"`
	Check  FGATestCheck `yaml:"check" json:"check"`
	Expect bool         `yaml:"expect" json:"expect"`
}

type FGATestTeardownSection struct {
	Warrants  []WarrantSetup `yaml:"warrants" json:"warrants"`
	Resources []string       `yaml:"resources" json:"resources"`
}

type FGATestYaml struct {
	Setup    FGATestSetupSection     `yaml:"setup" json:"setup"`
	Tests    []FGATestCase           `yaml:"tests" json:"tests"`
	Teardown *FGATestTeardownSection `yaml:"teardown,omitempty" json:"teardown,omitempty"`
}

type FGATestResult struct {
	Passed    bool
	NumPassed int
	NumTotal  int
}

func parseTestFile(testFile string) (FGATestYaml, error) {
	var testSpec FGATestYaml
	bytes, err := os.ReadFile(testFile)
	if err != nil {
		return testSpec, errors.Wrap(err, "failed to read test file")
	}

	if !isTestFile(testFile) {
		return testSpec, errors.Errorf("invalid test file format. %s", AcceptedTestFileExtensionErrorMessage)
	}

	switch {
	case strings.HasSuffix(testFile, TestFileExtensionYAML), strings.HasSuffix(testFile, TestFileExtensionYML):
		err = yaml.Unmarshal(bytes, &testSpec)
	case strings.HasSuffix(testFile, TestFileExtensionJSON):
		err = json.Unmarshal(bytes, &testSpec)
	default:
		return testSpec, errors.New("failed to parse test file: " + AcceptedTestFileExtensionErrorMessage)
	}
	if err != nil {
		return testSpec, errors.Wrap(err, "failed to parse test file")
	}
	// Check for duplicate test names
	testNames := make(map[string]bool)
	for _, t := range testSpec.Tests {
		if testNames[t.Name] {
			return testSpec, errors.Errorf("duplicate test name found: %s", t.Name)
		}
		testNames[t.Name] = true
	}
	return testSpec, nil
}

func setupTestWarrants(setup FGATestSetupSection) (string, error) {
	batchWrites := make([]fga.WriteWarrantOpts, 0, len(setup.Warrants))
	for _, w := range setup.Warrants {
		subjectType, subjectIdRelation, valid := strings.Cut(w.Subject, ResourceSeparator)
		if !valid {
			return "", errors.Errorf("invalid subject in setup: %s", w.Subject)
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, RelationSeparator)
		resourceType, resourceId, valid := strings.Cut(w.Resource, ResourceSeparator)
		if !valid {
			return "", errors.Errorf("invalid resource in setup: %s", w.Resource)
		}
		batchWrites = append(batchWrites, fga.WriteWarrantOpts{
			Op:           fga.WarrantOpCreate,
			ResourceType: resourceType,
			ResourceId:   resourceId,
			Relation:     w.Relation,
			Subject: fga.Subject{
				ResourceType: subjectType,
				ResourceId:   subjectId,
				Relation:     subjectRelation,
			},
			Policy: w.Policy,
		})
	}
	if len(batchWrites) > 0 {
		response, err := fga.BatchWriteWarrants(context.Background(), batchWrites)
		if err != nil {
			return "", errors.Wrap(err, "setup failed")
		}
		return response.WarrantToken, nil
	}
	return "", nil
}

func runTestCases(tests []FGATestCase, warrantToken string) (FGATestResult, error) {
	allPassed := true
	passed := 0
	total := len(tests)
	for _, t := range tests {
		subjectType, subjectIdRelation, valid := strings.Cut(t.Check.Subject, ResourceSeparator)
		if !valid {
			printer.PrintMsg(printer.RedText(printer.Cross, "Invalid subject in test: "+t.Check.Subject))
			allPassed = false
			continue
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, RelationSeparator)
		resourceType, resourceId, valid := strings.Cut(t.Check.Resource, ResourceSeparator)
		if !valid {
			printer.PrintMsg(printer.RedText(printer.Cross, "Invalid resource in test: "+t.Check.Resource))
			allPassed = false
			continue
		}
		warrantCheck := fga.WarrantCheck{
			ResourceType: resourceType,
			ResourceId:   resourceId,
			Relation:     t.Check.Relation,
			Subject: fga.Subject{
				ResourceType: subjectType,
				ResourceId:   subjectId,
				Relation:     subjectRelation,
			},
			Context: t.Check.Context,
		}
		result, err := fga.Check(
			context.Background(),
			fga.CheckOpts{
				Checks:       []fga.WarrantCheck{warrantCheck},
				WarrantToken: warrantToken,
			},
		)
		if err != nil {
			printer.PrintMsg(printer.RedText(printer.Cross, t.Name+": check failed: "+err.Error()))
			allPassed = false
			continue
		}
		if result.Authorized() == t.Expect {
			printer.PrintMsg(printer.GreenText(printer.Checkmark, t.Name))
			passed++
		} else {
			printer.PrintMsg(printer.RedText(printer.Cross, t.Name))
			allPassed = false
		}
	}
	return FGATestResult{Passed: allPassed, NumPassed: passed, NumTotal: total}, nil
}

func teardownTest(teardown *FGATestTeardownSection) error {
	if teardown == nil {
		return nil
	}
	for _, w := range teardown.Warrants {
		subjectType, subjectIdRelation, valid := strings.Cut(w.Subject, ResourceSeparator)
		if !valid {
			printer.PrintMsg(printer.RedText(printer.Cross, "Invalid subject in teardown: "+w.Subject))
			continue
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, RelationSeparator)
		resourceType, resourceId, valid := strings.Cut(w.Resource, ResourceSeparator)
		if !valid {
			printer.PrintMsg(printer.RedText(printer.Cross, "Invalid resource in teardown: "+w.Resource))
			continue
		}
		_, err := fga.WriteWarrant(
			context.Background(),
			fga.WriteWarrantOpts{
				Op:           fga.WarrantOpDelete,
				ResourceType: resourceType,
				ResourceId:   resourceId,
				Relation:     w.Relation,
				Subject: fga.Subject{
					ResourceType: subjectType,
					ResourceId:   subjectId,
					Relation:     subjectRelation,
				},
			},
		)
		if err != nil {
			printer.PrintMsg(printer.RedText(printer.Cross, "Teardown (warrant) failed: "+err.Error()))
		}
	}
	for _, resource := range teardown.Resources {
		resourceType, resourceId, valid := strings.Cut(resource, ResourceSeparator)
		if !valid {
			printer.PrintMsg(printer.RedText(printer.Cross, "Invalid resource in teardown: "+resource))
			continue
		}
		err := fga.DeleteResource(context.Background(), fga.DeleteResourceOpts{
			ResourceType: resourceType,
			ResourceId:   resourceId,
		})
		if err != nil {
			printer.PrintMsg(printer.RedText(printer.Cross, "Teardown (resource) failed: "+err.Error()))
		}
	}
	return nil
}

func cleanupTestResources(cmd *cobra.Command, setup FGATestSetupSection) error {
	cleanupResources, err := cmd.Flags().GetBool("cleanup")
	if err != nil {
		return errors.Wrap(err, "failed to get cleanup flag")
	}
	if cleanupResources {
		resourceSet := make(map[string]bool)
		for _, w := range setup.Warrants {
			resourceSet[w.Resource] = true
			resourceSet[w.Subject] = true
		}
		for resource := range resourceSet {
			resourceType, resourceId, valid := strings.Cut(resource, ResourceSeparator)
			if !valid {
				printer.PrintMsg(printer.RedText(printer.Cross, "Invalid resource in cleanup: "+resource))
				continue
			}
			err := fga.DeleteResource(context.Background(), fga.DeleteResourceOpts{
				ResourceType: resourceType,
				ResourceId:   resourceId,
			})
			if err != nil {
				printer.PrintMsg(printer.RedText(printer.Cross, "Cleanup (resource) failed: "+err.Error()))
			}
		}
	}
	return nil
}

func runTestFile(cmd *cobra.Command, testFile string) (FGATestResult, error) {
	testSpec, err := parseTestFile(testFile)
	if err != nil {
		return FGATestResult{}, err
	}
	warrantToken, err := setupTestWarrants(testSpec.Setup)
	if err != nil {
		printer.PrintMsg(printer.RedText(printer.Cross, err.Error()))
		return FGATestResult{}, err
	}
	result, err := runTestCases(testSpec.Tests, warrantToken)
	if err != nil {
		return result, err
	}
	if err := teardownTest(testSpec.Teardown); err != nil {
		return result, err
	}
	if err := cleanupTestResources(cmd, testSpec.Setup); err != nil {
		return result, err
	}
	return result, nil
}

func isTestFile(fileName string) bool {
	for _, extToCheck := range AcceptedTestFileExtensions {
		if strings.HasSuffix(fileName, extToCheck) {
			return true
		}
	}

	return false
}

func runTestDirectory(cmd *cobra.Command, path string) (FGATestResult, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return FGATestResult{}, errors.Wrap(err, "failed to read directory")
	}
	var testFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if isTestFile(name) {
			testFiles = append(testFiles, filepath.Join(path, name))
		}
	}
	if len(testFiles) == 0 {
		printer.PrintMsg("No test files found in directory. " + AcceptedTestFileExtensionErrorMessage)
		return FGATestResult{Passed: true, NumPassed: 0, NumTotal: 0}, nil
	}

	finalResult := FGATestResult{Passed: true, NumPassed: 0, NumTotal: 0}
	for _, file := range testFiles {
		printer.PrintMsg("")
		printer.PrintMsg("Running tests in " + file)
		result, err := runTestFile(cmd, file)

		finalResult.NumPassed += result.NumPassed
		finalResult.NumTotal += result.NumTotal

		if err != nil {
			printer.PrintMsg(printer.RedText(printer.Cross, err.Error()))
			finalResult.Passed = false
		} else if !result.Passed {
			finalResult.Passed = false
		}
	}

	printer.PrintMsg("")
	if finalResult.Passed {
		printer.PrintMsg(fmt.Sprintf("All tests passed (%d/%d)", finalResult.NumPassed, finalResult.NumTotal))
		return finalResult, nil
	} else {
		printer.PrintMsg(fmt.Sprintf("Some tests failed (%d/%d)", finalResult.NumPassed, finalResult.NumTotal))
		return finalResult, errors.New("some tests failed")
	}
}

var testCmd = &cobra.Command{
	Use:   "test <test_file.test.yaml|json|directory>",
	Short: "Run FGA permission tests from a YAML/JSON file or all test files in a directory",
	Long: `Run FGA permission tests from a YAML or JSON file, or all .test.yaml/.test.json files in a directory, setting up and tearing down warrants and resources as specified.

Test file format (YAML or JSON):

YAML:
  setup:
    warrants:
      - subject: user:admin
        relation: owner
        resource: doc:1
  tests:
    - name: Admin is owner
      check:
        subject: user:admin
        relation: owner
        resource: doc:1
        context:
          key: "some_value"
      expect: true
  
  # Optional teardown section, alternatively pass the --cleanup flag to the test command to clean up all resources from the setup section
  teardown:
    resources:
      - doc:1
      - user:admin
    warrants:
      - subject: user:admin
        relation: owner
        resource: doc:1

JSON:
{
  "setup": {
    "warrants": [
      { "subject": "user:admin", "relation": "owner", "resource": "doc:1" }
    ]
  },
  "tests": [
    {
      "name": "Admin is owner",
      "check": { "subject": "user:admin", "relation": "owner", "resource": "doc:1", "context": { "key": "some_value" } },
      "expect": true
    }
  ],
}

Each test file must have a 'setup' section, a 'tests' array, and an optional 'teardown' section.
`,
	Example: "workos fga test ./schema.test.yaml --cleanup\nworkos fga test ./schema.test.json --cleanup\nworkos fga test ./test-directory/ --cleanup",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		info, err := os.Stat(path)
		if err != nil {
			return errors.Wrap(err, "failed to stat path")
		}
		if info.IsDir() {
			result, err := runTestDirectory(cmd, path)
			if err != nil {
				return err
			}
			if result.Passed {
				return nil
			} else {
				return errors.New("some tests failed")
			}
		} else {
			result, err := runTestFile(cmd, path)
			printer.PrintMsg("")
			if err != nil {
				printer.PrintMsg(printer.RedText(printer.Cross, err.Error()))
				return err
			}
			if result.Passed {
				printer.PrintMsg(fmt.Sprintf("All tests passed (%d/%d)", result.NumPassed, result.NumTotal))
				return nil
			} else {
				printer.PrintMsg(fmt.Sprintf("Some tests failed (%d/%d)", result.NumPassed, result.NumTotal))
				return errors.New("some tests failed")
			}
		}
	},
}

func convertSchemaError(err error) error {
	var target workos_errors.HTTPError
	if errors.As(err, &target) {
		formattedErrors := make([]string, 0, len(target.Errors))
		for _, e := range target.Errors {
			formattedErrors = append(formattedErrors, printer.RedText("\t* "+e))
		}
		if len(target.Errors) > 0 {
			return errors.New(printer.RedText(fmt.Sprintf("error validating schema: %s\n%s", target.Message, strings.Join(formattedErrors, "\n"))))
		}
	}
	return errors.New(printer.RedText(fmt.Sprintf("error validating schema: %v", err)))
}

func warrantCheckAsString(w fga.WarrantCheck) (string, error) {
	s := fmt.Sprintf(
		"%s:%s %s %s:%s",
		w.Subject.ResourceType,
		w.Subject.ResourceId,
		w.Relation,
		w.ResourceType,
		w.ResourceId,
	)
	if len(w.Context) > 0 {
		bytes, err := json.Marshal(w.Context)
		if err != nil {
			return "", err
		}
		s = fmt.Sprintf("%s '%s'", s, string(bytes))
	}

	return s, nil
}

func buildDecisionTreeList(node fga.DecisionTreeNode) *list.List {
	checkText := fmt.Sprintf(
		"%s:%s#%s@%s:%s",
		node.Check.ResourceType,
		node.Check.ResourceId,
		node.Check.Relation,
		node.Check.Subject.ResourceType,
		node.Check.Subject.ResourceId,
	)
	if node.Check.Subject.Relation != "" {
		checkText = fmt.Sprintf("%s#%s", checkText, node.Check.Subject.Relation)
	}
	if node.Policy != "" {
		checkText = fmt.Sprintf("%s - %s", checkText, node.Policy)
	}

	// Prefix with check mark (matched), cross (not_matched), or question mark (eval_policy)
	switch node.Decision {
	case "matched":
		checkText = fmt.Sprintf("%s %s", printer.GreenText(printer.Checkmark), checkText)
	case "not_matched":
		checkText = fmt.Sprintf("%s %s", printer.RedText(printer.Cross), checkText)
	case "eval_policy":
		checkText = fmt.Sprintf("%s %s", printer.YellowText(printer.QuestionMark), checkText)
	}

	checkText = fmt.Sprintf("%s (%dms)", checkText, node.ProcessingTime/1000000)
	tree := list.New(checkText).Enumerator(list.Tree)
	for _, child := range node.Children {
		tree.Item(buildDecisionTreeList(child))
	}

	return tree
}
