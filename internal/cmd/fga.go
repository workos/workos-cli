package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/workos/workos-cli/internal/printer"
	"github.com/workos/workos-go/v4/pkg/fga"
)

var objectTypesFile string

func init() {
	// object-types
	listObjectTypesCmd.Flags().Int("limit", 10, "limit the number of results returned")
	listObjectTypesCmd.Flags().String("before", "", "cursor indicating results that occur before a specific result")
	listObjectTypesCmd.Flags().String("after", "", "cursor indicating results that occur after a specific result")
	listObjectTypesCmd.Flags().String("order", "", "order in which a list of results should be returned (asc or desc)")
	objectTypeCmd.AddCommand(listObjectTypesCmd)
	applyObjectTypesCmd.Flags().StringVarP(&objectTypesFile, "file", "f", "", "file containing object type definitions")
	objectTypeCmd.AddCommand(applyObjectTypesCmd)
	fgaCmd.AddCommand(objectTypeCmd)

	// warrants
	fgaCmd.AddCommand(assignRelationCmd)
	fgaCmd.AddCommand(removeRelationCmd)

	// check
	checkRelationCmd.Flags().StringP("warrantToken", "w", "", "warrant token to use for check")
	checkRelationCmd.Flags().String("assert", "", "assert that the check is true or false")
	checkRelationCmd.Flags().BoolP("debug", "d", false, "run check in debug mode")
	fgaCmd.AddCommand(checkRelationCmd)

	// objects
	objectCmd.AddCommand(createObjectCmd)
	listObjectsCmd.Flags().String("type", "", "object type to filter results by")
	listObjectsCmd.Flags().String("search", "", "search term to filter a list of results by")
	listObjectsCmd.Flags().Int("limit", 10, "limit the number of results returned")
	listObjectsCmd.Flags().String("before", "", "cursor indicating results that occur before a specific result")
	listObjectsCmd.Flags().String("after", "", "cursor indicating results that occur after a specific result")
	listObjectsCmd.Flags().String("order", "", "order in which a list of results should be returned (asc or desc)")
	objectCmd.AddCommand(listObjectsCmd)
	objectCmd.AddCommand(updateObjectCmd)
	objectCmd.AddCommand(deleteObjectCmd)
	fgaCmd.AddCommand(objectCmd)

	// query
	queryCmd.Flags().StringP("warrantToken", "w", "", "warrant token to use for query")
	queryCmd.Flags().Int("limit", 10, "limit the number of results returned")
	queryCmd.Flags().String("before", "", "cursor indicating results that occur before a specific result")
	queryCmd.Flags().String("after", "", "cursor indicating results that occur after a specific result")
	queryCmd.Flags().String("order", "", "order in which a list of results should be returned (asc or desc)")
	fgaCmd.AddCommand(queryCmd)

	rootCmd.AddCommand(fgaCmd)
}

var fgaCmd = &cobra.Command{
	Use:   "fga",
	Short: "Manage FGA resources (object types, warrants, and objects).",
	Long:  "Manage FGA-specific resources like object types, warrants, and objects and perform check and query operations to validate your FGA model.",
}

var objectTypeCmd = &cobra.Command{
	Use:   "objecttype",
	Short: "Manage your object types",
	Long:  "List and apply object types. Object types are used to define the types of objects in your system and the relations between them.",
}

var listObjectTypesCmd = &cobra.Command{
	Use:     "list",
	Short:   "List object types",
	Long:    "List object types, optionally providing common flags to filter and paginate the results.",
	Example: "workos fga objecttype list --limit=5",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		objectTypes, err := fga.ListObjectTypes(context.Background(), fga.ListObjectTypesOpts{
			Limit: 100,
		})
		if err != nil {
			return fmt.Errorf("error listing object types: %v", err)
		}

		tbl := printer.NewTable().Headers(
			printer.TableHeader("Type"),
			printer.TableHeader("Definition"),
		)
		for _, objectType := range objectTypes.Data {
			definitionString, err := json.MarshalIndent(objectType, "", "    ")
			if err != nil {
				return fmt.Errorf("error listing object types: %v", err)
			}
			tbl.Row(
				objectType.Type,
				string(definitionString),
			)
		}

		fmt.Println(tbl.Render())
		fmt.Printf("before: %s\n", objectTypes.ListMetadata.Before)
		fmt.Printf("after: %s\n", objectTypes.ListMetadata.After)
		return nil
	},
}

var applyObjectTypesCmd = &cobra.Command{
	Use:     "apply",
	Short:   "Apply a set of object types",
	Long:    "Apply a set of object types from a specified file. This command will create any object types present in the file and delete any object types that are not.",
	Example: "workos fga objecttype apply -f object-types.json",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var bytes []byte
		var err error
		if objectTypesFile != "" {
			jsonFile, err := os.Open(objectTypesFile)
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

		var objectTypes []fga.UpdateObjectTypeOpts
		err = json.Unmarshal(bytes, &objectTypes)
		if err != nil {
			return err
		}

		_, err = fga.BatchUpdateObjectTypes(context.Background(), objectTypes)
		if err != nil {
			return err
		}

		fmt.Println("object types updated")
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

		fmt.Printf("assigned %s %s %s\nWarrant-Token: %s\n", args[0], args[1], args[2], res.WarrantToken)
		return nil
	},
}

var removeRelationCmd = &cobra.Command{
	Use:     "remove <subject> <relation> <object>",
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

		fmt.Printf("removed %s %s %s\nwarrant_token: %s\n", args[0], args[1], args[2], res.WarrantToken)
		return nil
	},
}

var objectCmd = &cobra.Command{
	Use:   "object",
	Short: "Manage your objects",
	Long:  "Create, update, read, list and delete objects.",
}

var createObjectCmd = &cobra.Command{
	Use:     "create <object> [meta]",
	Short:   "Create a new object",
	Long:    "Create a new object of a given type, optionally providing an identifier for the object and/or any metadata to attach to the object.",
	Example: `workos fga object create user:john '{"email":"john.doe@workos.com"}'`,
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		objectType, objectId, valid := strings.Cut(args[0], ":")
		if !valid {
			return fmt.Errorf("invalid object: %s", args[0])
		}

		var meta map[string]interface{}
		var err error
		if len(args) == 2 {
			err = json.Unmarshal([]byte(args[1]), &meta)
			if err != nil {
				return fmt.Errorf("invalid object meta: %s", args[1])
			}
		}

		createdObject, err := fga.CreateObject(context.Background(), fga.CreateObjectOpts{
			ObjectType: objectType,
			ObjectId:   objectId,
			Meta:       meta,
		})
		if err != nil {
			return fmt.Errorf("error creating object: %v", err)
		}

		if len(createdObject.Meta) > 0 {
			fmt.Printf("created object %s:%s (%v)\n", createdObject.ObjectType, createdObject.ObjectId, createdObject.Meta)
		} else {
			fmt.Printf("created object %s:%s\n", createdObject.ObjectType, createdObject.ObjectId)
		}

		return nil
	},
}

var listObjectsCmd = &cobra.Command{
	Use:     "list",
	Short:   "List objects",
	Long:    "List objects, optionally specifying the '--type' flag to filter to objects of a specific type or providing common flags to filter and paginate the results.",
	Example: "workos fga object list --type=user --limit=15",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		objectType, err := cmd.Flags().GetString("type")
		if err != nil {
			return fmt.Errorf("invalid type flag")
		}
		search, err := cmd.Flags().GetString("search")
		if err != nil {
			return fmt.Errorf("invalid search flag")
		}
		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return fmt.Errorf("invalid limit flag")
		}
		before, err := cmd.Flags().GetString("before")
		if err != nil {
			return fmt.Errorf("invalid before flag")
		}
		after, err := cmd.Flags().GetString("after")
		if err != nil {
			return fmt.Errorf("invalid after flag")
		}
		order, err := cmd.Flags().GetString("order")
		if err != nil {
			return fmt.Errorf("invalid order flag")
		}
		var orderFilter fga.Order
		if order != "" {
			if strings.ToLower(order) == "asc" {
				orderFilter = fga.Asc
			} else {
				orderFilter = fga.Desc
			}
		}

		objects, err := fga.ListObjects(context.Background(), fga.ListObjectsOpts{
			ObjectType: objectType,
			Search:     search,
			Limit:      limit,
			Before:     before,
			After:      after,
			Order:      orderFilter,
		})
		if err != nil {
			return fmt.Errorf("error listing objects: %v", err)
		}

		tbl := printer.NewTable().Headers(
			printer.TableHeader("Object Type"),
			printer.TableHeader("Object ID"),
			printer.TableHeader("Meta"),
		)
		for _, object := range objects.Data {
			metaString, err := json.MarshalIndent(object.Meta, "", "    ")
			if err != nil {
				return fmt.Errorf("error listing objects: %v", err)
			}
			tbl.Row(
				object.ObjectType,
				object.ObjectId,
				string(metaString),
			)
		}

		fmt.Println(tbl.Render())
		fmt.Printf("before: %s\n", objects.ListMetadata.Before)
		fmt.Printf("after: %s\n", objects.ListMetadata.After)
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
		objectType, objectId, valid := strings.Cut(args[0], ":")
		if !valid {
			return fmt.Errorf("invalid object: %s", args[0])
		}

		var meta map[string]interface{}
		err := json.Unmarshal([]byte(args[1]), &meta)
		if err != nil {
			return fmt.Errorf("invalid meta: %s", args[1])
		}

		updatedObject, err := fga.UpdateObject(context.Background(), fga.UpdateObjectOpts{
			ObjectType: objectType,
			ObjectId:   objectId,
			Meta:       meta,
		})
		if err != nil {
			return fmt.Errorf("error updating object: %v", err)
		}

		fmt.Printf("updated object %s:%s\n", updatedObject.ObjectType, updatedObject.ObjectId)
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
		objectType, objectId, valid := strings.Cut(args[0], ":")
		if !valid {
			return fmt.Errorf("invalid object: %s", args[0])
		}

		err := fga.DeleteObject(context.Background(), fga.DeleteObjectOpts{
			ObjectType: objectType,
			ObjectId:   objectId,
		})
		if err != nil {
			return fmt.Errorf("error deleting object: %v", err)
		}

		fmt.Printf("deleted object %s\n", args[0])
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

		var policyContext map[string]interface{}
		if len(args) > 3 {
			err := json.Unmarshal([]byte(args[3]), &policyContext)
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
			Context: policyContext,
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
			return fmt.Errorf("error evaluating check: %v", err)
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
				fmt.Printf("%s %s\n", printer.GreenText(printer.Checkmark, fmt.Sprintf("assert %t", assertBool)), warrantCheckString)
			} else {
				fmt.Printf("%s %s\n", printer.RedText(printer.Cross, fmt.Sprintf("assert %t", assertBool)), warrantCheckString)
				os.Exit(1)
			}
		} else if result.Authorized() {
			fmt.Printf("%s %s\n", printer.GreenText(printer.Checkmark, "true"), warrantCheckString)
		} else {
			fmt.Printf("%s %s\n", printer.RedText(printer.Cross, "false"), warrantCheckString)
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
		after, _ := cmd.Flags().GetString("after")
		before, _ := cmd.Flags().GetString("before")
		limit, _ := cmd.Flags().GetInt("limit")
		order, _ := cmd.Flags().GetString("order")
		warrantToken, _ := cmd.Flags().GetString("warrantToken")
		var policyContext map[string]interface{}
		if len(args) > 1 {
			err := json.Unmarshal([]byte(args[3]), &policyContext)
			if err != nil {
				return fmt.Errorf("invalid context: %s", args[3])
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
			return fmt.Errorf("error performing query: %v", err)
		}

		tbl := printer.NewTable().Headers(
			printer.TableHeader("Object Type"),
			printer.TableHeader("Object ID"),
			printer.TableHeader("Relation"),
			printer.TableHeader("Implicit"),
			printer.TableHeader("Meta"),
		)
		for _, queryResult := range result.Data {
			metaString, err := json.MarshalIndent(queryResult.Meta, "", "    ")
			if err != nil {
				return fmt.Errorf("error listing objects: %v", err)
			}
			tbl.Row(
				queryResult.ObjectType,
				queryResult.ObjectId,
				queryResult.Relation,
				strconv.FormatBool(queryResult.IsImplicit),
				string(metaString),
			)
		}

		fmt.Println(tbl.Render())
		fmt.Printf("before: %s\n", result.ListMetadata.Before)
		fmt.Printf("after: %s\n", result.ListMetadata.After)
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
