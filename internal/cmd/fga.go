package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/printer"
	"github.com/workos/workos-go/v4/pkg/fga"
	"os"
	"strconv"
	"strings"
)

func init() {
	// object-types
	fgaCmd.AddCommand(listObjectTypesCmd)
	fgaCmd.AddCommand(applyObjectTypesCmd)

	// warrants
	fgaCmd.AddCommand(assignRelationCmd)
	fgaCmd.AddCommand(removeRelationCmd)

	// check
	fgaCmd.AddCommand(checkRelationCmd)
	checkRelationCmd.Flags().StringP("warrantToken", "w", "", "warrant token to use for check")
	checkRelationCmd.Flags().String("assert", "", "assert that the check is true or false")
	checkRelationCmd.Flags().BoolP("debug", "d", false, "run check in debug mode")

	// objects
	fgaCmd.AddCommand(createObjectCmd)
	fgaCmd.AddCommand(listObjectsCmd)
	listObjectsCmd.Flags().String("type", "", "object type to filter results by")
	listObjectsCmd.Flags().String("search", "", "search term to filter a list of results by")
	listObjectsCmd.Flags().Int("limit", 10, "limit the number of results returned")
	listObjectsCmd.Flags().String("before", "", "cursor indicating results that occur before a specific result")
	listObjectsCmd.Flags().String("after", "", "cursor indicating results that occur after a specific result")
	listObjectsCmd.Flags().String("order", "", "order in which a list of results should be returned (asc or desc)")
	fgaCmd.AddCommand(updateObjectCmd)
	fgaCmd.AddCommand(deleteObjectCmd)

	// query
	fgaCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringP("warrantToken", "w", "", "warrant token to use for query")
	queryCmd.Flags().Int("limit", 10, "limit the number of results returned")
	queryCmd.Flags().String("before", "", "cursor indicating results that occur before a specific result")
	queryCmd.Flags().String("after", "", "cursor indicating results that occur after a specific result")
	queryCmd.Flags().String("order", "", "order in which a list of results should be returned (asc or desc)")

	rootCmd.AddCommand(fgaCmd)
}

var fgaCmd = &cobra.Command{
	Use:   "fga",
	Short: "Manage FGA resources (object types, warrants, and objects).",
	Long:  "Manage FGA-specific resources like object types, warrants, and objects and perform check and query operations to validate your FGA model.",
}

var listObjectTypesCmd = &cobra.Command{
	Use:     "objecttype list",
	Short:   "List object types",
	Long:    "List object types, optionally providing common flags to filter and paginate the results.",
	Example: "workos fga objecttype list --limit=5",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var applyObjectTypesCmd = &cobra.Command{
	Use:     "objecttype apply",
	Short:   "Apply a set of object types",
	Long:    "Apply a set of object types from a specified file. This command will create any object types present in the file and delete any object types that are not.",
	Example: "workos fga objecttype apply -f object-types.json",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var assignRelationCmd = &cobra.Command{
	Use:     "assign <subject> <relation> <object> [policy]",
	Short:   "Assign a relation",
	Long:    "Assign a relation between a given subject and a given object, optionally specifying a policy that dictates when the relation applies.",
	Example: "workos fga assign user:john owner document:xyz",
	Args:    cobra.RangeArgs(3, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectType, subjectIdRelation, valid := strings.Cut(args[0], ":")
		if !valid {
			return fmt.Errorf("invalid subject: %s", args[0])
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, "#")
		relation := args[1]
		objectType, objectId, valid := strings.Cut(args[2], ":")
		if !valid {
			return fmt.Errorf("invalid object: %s", args[0])
		}

		var policy string
		if len(args) > 3 {
			policy = args[3]
		}

		res, err := fga.WriteWarrant(
			context.Background(),
			fga.WriteWarrantOpts{
				Op:         "create",
				ObjectType: objectType,
				ObjectId:   objectId,
				Relation:   relation,
				Subject: fga.Subject{
					ObjectType: subjectType,
					ObjectId:   subjectId,
					Relation:   subjectRelation,
				},
				Policy: policy,
			},
		)
		if err != nil {
			return fmt.Errorf("error assigning relation: %v", err)
		}

		fmt.Printf("%s assigned %s on %s:\nWarrant-Token: %s\n", args[0], args[1], args[2], res.WarrantToken)
		return nil
	},
}

var removeRelationCmd = &cobra.Command{
	Use:     "assign <subject> <relation> <object>",
	Short:   "Remove a relation",
	Long:    "Remove a relation between a given subject and a given object.",
	Example: "workos fga remove user:john owner document:xyz",
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectType, subjectIdRelation, valid := strings.Cut(args[0], ":")
		if !valid {
			return fmt.Errorf("invalid subject: %s", args[0])
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, "#")
		relation := args[1]
		objectType, objectId, valid := strings.Cut(args[2], ":")
		if !valid {
			return fmt.Errorf("invalid object: %s", args[0])
		}

		res, err := fga.WriteWarrant(
			context.Background(),
			fga.WriteWarrantOpts{
				Op:         "delete",
				ObjectType: objectType,
				ObjectId:   objectId,
				Relation:   relation,
				Subject: fga.Subject{
					ObjectType: subjectType,
					ObjectId:   subjectId,
					Relation:   subjectRelation,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("error removing relation: %v", err)
		}

		fmt.Printf("removed %s from %s on %s:\nwarrant_token: %s\n", args[1], args[0], args[2], res.WarrantToken)
		return nil
	},
}

var createObjectCmd = &cobra.Command{
	Use:     "create <object> [meta]",
	Short:   "Create a new object",
	Long:    "Create a new object of a given type, optionally providing an identifier for the object and/or any metadata to attach to the object.",
	Example: `workos fga object create user:john '{"email":"john.doe@workos.com"}'`,
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var listObjectsCmd = &cobra.Command{
	Use:     "object list",
	Short:   "List objects",
	Long:    "List objects, optionally specifying the '--objectType' flag to filter to objects of a specific type or providing common flags to filter and paginate the results.",
	Example: "workos fga object list --objectType=user --limit=15",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		//limit, _ := cmd.Flags().GetInt("limit")
		//before, _ := cmd.Flags().GetString("before")
		//after, _ := cmd.Flags().GetString("after")
		//order, _ := cmd.Flags().GetString("order")
		return nil
	},
}

var updateObjectCmd = &cobra.Command{
	Use:     "update <object> <meta>",
	Short:   "Update an object",
	Long:    "Update an object, providing metadata to attach to it.",
	Example: `workos fga object update user:john '{"email":"john.doe@workos.com"}'`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var deleteObjectCmd = &cobra.Command{
	Use:     "delete <object>",
	Short:   "Delete an object",
	Long:    "Delete a given object. This will delete any warrants associated with the object.",
	Example: `workos fga object delete user:john`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var checkRelationCmd = &cobra.Command{
	Use:     "check <subject> <relation> <object> [context]",
	Short:   "Check for a relation",
	Long:    "Check if a given subject has the specified relation on a given object, optionally specifying context to use while evaluating the check.",
	Example: `workos fga check user:john owner document:xyz '{"organization": "acme"}'`,
	Args:    cobra.RangeArgs(3, 4),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectType, subjectIdRelation, valid := strings.Cut(args[0], ":")
		if !valid {
			return fmt.Errorf("invalid subject: %s", args[0])
		}
		subjectId, subjectRelation, _ := strings.Cut(subjectIdRelation, "#")
		relation := args[1]
		objectType, objectId, valid := strings.Cut(args[2], ":")
		if !valid {
			return fmt.Errorf("invalid object: %s", args[0])
		}

		var checkContext map[string]interface{}
		if len(args) > 3 {
			err := json.Unmarshal([]byte(args[3]), &checkContext)
			if err != nil {
				return fmt.Errorf("invalid context: %s", args[3])
			}
		}

		warrantToken, _ := cmd.Flags().GetString("warrantToken")
		debug, _ := cmd.Flags().GetBool("debug")

		warrantCheck := fga.WarrantCheck{
			ObjectType: objectType,
			ObjectId:   objectId,
			Relation:   relation,
			Subject: fga.Subject{
				ObjectType: subjectType,
				ObjectId:   subjectId,
				Relation:   subjectRelation,
			},
			Context: checkContext,
		}
		result, err := fga.Check(
			context.Background(),
			fga.CheckOpts{
				Warrant:      warrantCheck,
				WarrantToken: warrantToken,
				Debug:        debug,
			},
		)
		if err != nil {
			return fmt.Errorf("error performing check: %v", err)
		}

		warrantCheckString, err := warrantCheckAsString(warrantCheck)
		if err != nil {
			return fmt.Errorf("invalid check: %v", err)
		}

		assert, _ := cmd.Flags().GetString("assert")
		if assert != "" {
			assertBool, err := strconv.ParseBool(assert)
			if err != nil {
				return fmt.Errorf("invalid assertion: %s", assert)
			}

			if assertBool == result.Authorized() {
				fmt.Printf("%s %s\n", termenv.String(printer.Checkmark, fmt.Sprintf("assert %t", assertBool)).Foreground(printer.Green), warrantCheckString)
			} else {
				fmt.Printf("%s %s\n", termenv.String(printer.Cross, fmt.Sprintf("assert %t", assertBool)).Foreground(printer.Red), warrantCheckString)
				os.Exit(1)
			}
		} else if result.Authorized() {
			fmt.Printf("%s %s\n", termenv.String(printer.Checkmark, "true").Foreground(printer.Green), warrantCheckString)
		} else {
			fmt.Printf("%s %s\n", termenv.String(printer.Cross, "false").Foreground(printer.Red), warrantCheckString)
		}

		return nil
	},
}

var queryCmd = &cobra.Command{
	Use:     "query <query> [context]",
	Short:   "Query for access rules",
	Long:    "Run a query to see which objects a subject has access to or which subjects have access to an object, optionally specifying context to use while evaluating the query.",
	Example: "workos fga query select document where user:john is owner",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func warrantCheckAsString(w fga.WarrantCheck) (string, error) {
	s := fmt.Sprintf(
		"%s:%s %s %s:%s",
		w.Subject.ObjectType,
		w.Subject.ObjectId,
		w.Relation,
		w.ObjectType,
		w.ObjectId,
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
