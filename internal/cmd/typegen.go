package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TypeGenerator struct {
	writer io.Writer
}

func NewTypeGenerator(w io.Writer) *TypeGenerator {
	return &TypeGenerator{writer: w}
}

// Permission types
type PermissionResponse struct {
	Data []Permission `json:"data"`
}

type Permission struct {
	Slug string `json:"slug"`
}

// Permission Generator
func (g *TypeGenerator) GenerateFromPermissionsAPI() error {
	jsonData, err := os.ReadFile("permissions.json")
	if err != nil {
		return err
	}

	var response PermissionResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return fmt.Errorf("failed to parse permissions JSON: %w", err)
	}

	slugs := make([]string, len(response.Data))
	for i, permission := range response.Data {
		slugs[i] = permission.Slug
	}

	return g.generateUnionType("WorkOSPermission", slugs)
}

// Role types
type RoleResponse struct {
	Data []Role `json:"data"`
}

type Role struct {
	Slug string `json:"slug"`
}

// Role Generator
func (g *TypeGenerator) GenerateFromRolesAPI() error {
	jsonData, err := os.ReadFile("roles.json")
	if err != nil {
		return err
	}

	var response RoleResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return fmt.Errorf("failed to parse roles JSON: %w", err)
	}

	slugs := make([]string, len(response.Data))
	for i, role := range response.Data {
		slugs[i] = role.Slug
	}

	return g.generateUnionType("WorkOSRole", slugs)
}

// Audit Log types
type AuditLogResponse struct {
	Data []AuditLogEvent `json:"data"`
}

type AuditLogEvent struct {
	Name   string         `json:"name"`
	Schema AuditLogSchema `json:"schema"`
}

type AuditLogSchema struct {
	ActorSchema    AuditLogActorSchema     `json:"actor"`
	TargetSchemas  []AuditLogTargetSchema  `json:"targets"`
	MetadataSchema *AuditLogMetadataSchema `json:"metadata"`
}

type AuditLogActorSchema struct {
	MetadataSchema *AuditLogMetadataSchema `json:"metadata"`
}

type AuditLogTargetSchema struct {
	Type           string                  `json:"type"`
	MetadataSchema *AuditLogMetadataSchema `json:"metadata"`
}

type AuditLogMetadataSchema struct {
	Type       string                              `json:"type"`
	Properties map[string]AuditLogMetadataProperty `json:"properties"`
}

type AuditLogMetadataProperty struct {
	Type     string `json:"type"`
	Nullable *bool  `json:"nullable,omitempty"`
}

// Audit Log Generator
func (g *TypeGenerator) GenerateFromAuditLogsAPI() error {
	req, err := http.NewRequest("GET", "https://api.workos.com/audit_logs/actions", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("WORKOS_API_KEY"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch audit logs actions: %w", err)
	}
	defer resp.Body.Close()

	var response AuditLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to parse audit logs response: %w", err)
	}

	var interfaces []string
	var eventNames []string

	for _, action := range response.Data {
		interfaceName := g.generateInterfaceName(action.Name) + "AuditLogEvent"
		eventNames = append(eventNames, interfaceName)

		interfaceContent := g.generateAuditLogEventInterface(action)
		interfaces = append(interfaces, interfaceContent)
	}

	output := strings.Join(interfaces, "\n\n")

	unionType := fmt.Sprintf("\nexport type AuditLogEvent = %s;\n",
		strings.Join(eventNames, " | "))
	output += unionType

	_, err = fmt.Fprint(g.writer, output)
	return err
}

func (g *TypeGenerator) generateAuditLogEventInterface(event AuditLogEvent) string {
	interfaceName := g.generateInterfaceName(event.Name)

	var properties []string

	properties = append(properties, fmt.Sprintf("  action: '%s';", event.Name))
	properties = append(properties, "  occurredAt: string;")
	properties = append(properties, fmt.Sprintf("  version?: number;"))

	actorProps := []string{
		"    id: string;",
		"    name?: string;",
	}

	if event.Schema.ActorSchema.MetadataSchema != nil {
		metadataProps := g.generateMetadataProperties(event.Schema.ActorSchema.MetadataSchema.Properties)
		actorProps = append(actorProps, fmt.Sprintf("    metadata: {\n%s\n    };", metadataProps))
	}

	properties = append(properties, fmt.Sprintf("  actor: {\n%s\n  };", strings.Join(actorProps, "\n")))

	if len(event.Schema.TargetSchemas) > 0 {
		var targetTypes []string
		for _, target := range event.Schema.TargetSchemas {
			targetProps := []string{
				fmt.Sprintf("    type: '%s';", target.Type),
				"    id: string;",
				"    name?: string;",
			}
			if target.MetadataSchema != nil {
				metadataProps := g.generateMetadataProperties(target.MetadataSchema.Properties)
				targetProps = append(targetProps, fmt.Sprintf("    metadata: {\n%s\n    };", metadataProps))
			}

			targetTypes = append(targetTypes, fmt.Sprintf("{\n%s\n}", strings.Join(targetProps, "\n")))
		}

		properties = append(properties, fmt.Sprintf("  targets: (%s)[];", strings.Join(targetTypes, " | ")))
	}

	contextProps := []string{
		"    location: string;",
		"    userAgent?: string;",
	}

	properties = append(properties, fmt.Sprintf("  context: {\n%s\n  };", strings.Join(contextProps, "\n")))

	if event.Schema.MetadataSchema != nil {
		metadataProps := g.generateMetadataProperties(event.Schema.MetadataSchema.Properties)
		properties = append(properties, fmt.Sprintf("  metadata: {\n%s\n  };", metadataProps))
	}

	return fmt.Sprintf("export interface %s {\n%s\n}",
		interfaceName,
		strings.Join(properties, "\n"))
}

func (g *TypeGenerator) generateMetadataProperties(properties map[string]AuditLogMetadataProperty) string {
	var props []string
	for key, prop := range properties {
		nullable := ""
		if prop.Nullable != nil && *prop.Nullable {
			nullable = " | null"
		}
		props = append(props, fmt.Sprintf("      %s: %s%s;", key, prop.Type, nullable))
	}

	return strings.Join(props, "\n")
}

// Utils
func (g *TypeGenerator) generateInterfaceName(action string) string {
	parts := strings.Split(action, ".")
	var name string
	for _, part := range parts {
		part = strings.ReplaceAll(part, "_", " ")
		words := strings.Fields(part)
		for _, word := range words {
			name += cases.Title(language.English).String(word)
		}
	}

	return name
}

func (g *TypeGenerator) generateUnionType(typeName string, values []string) error {
	quotedValues := make([]string, len(values))
	for i, v := range values {
		quotedValues[i] = fmt.Sprintf("'%s'", v)
	}

	output := fmt.Sprintf("export type %s = %s;\n", typeName, strings.Join(quotedValues, " | "))
	_, err := fmt.Fprint(g.writer, output)

	return err
}

// CLI
var typegenCmd = &cobra.Command{
	Use:   "typegen",
	Short: "Generate TypeScript types from WorkOS APIs",
	Long:  `A tool to generate TypeScript type definitions from WorkOS permissions, roles, and audit logs.`,
}

func NewTypeGeneratorCmd() *cobra.Command {
	cmd := typegenCmd

	cmd.AddCommand(newPermissionsCmd())
	cmd.AddCommand(newRolesCmd())
	cmd.AddCommand(newAuditLogsCmd())

	return cmd
}

func newPermissionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "permissions",
		Short: "Generate TypeScript types for WorkOS permissions",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator := NewTypeGenerator(cmd.OutOrStdout())
			return generator.GenerateFromPermissionsAPI()
		},
	}
}

func newRolesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "roles",
		Short: "Generate TypeScript types for WorkOS roles",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator := NewTypeGenerator(cmd.OutOrStdout())
			return generator.GenerateFromRolesAPI()
		},
	}
}

func newAuditLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "audit-logs",
		Short: "Generate TypeScript types for WorkOS audit logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			generator := NewTypeGenerator(cmd.OutOrStdout())
			return generator.GenerateFromAuditLogsAPI()
		},
	}
}

func init() {
	cmd := NewTypeGeneratorCmd()
	rootCmd.AddCommand(cmd)
}
