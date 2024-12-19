package typegen

import (
	"fmt"
	"io"
	"strings"

	"github.com/workos/workos-go/v4/pkg/auditlogs"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TypeScriptTypeGenerator struct {
	writer io.Writer
}

func NewTypeScriptTypeGenerator(w io.Writer) *TypeScriptTypeGenerator {
	return &TypeScriptTypeGenerator{writer: w}
}

// Import generator
func (g *TypeScriptTypeGenerator) generateImports(resources []string) error {
	// No imports needed for TypeScript
	return nil
}

// Audit Log Generator
func (g *TypeScriptTypeGenerator) generateAuditLogActionsTypes(actions []auditlogs.AuditLogAction) error {
	var interfaces []string
	var eventNames []string

	for _, action := range actions {
		interfaceName := g.generateInterfaceName(action.Name)
		eventNames = append(eventNames, interfaceName)

		interfaceContent := g.generateAuditLogEventInterface(action)
		interfaces = append(interfaces, interfaceContent)
	}

	output := strings.Join(interfaces, "\n")

	unionType := fmt.Sprintf("\nexport type AuditLogEvent = %s;\n",
		strings.Join(eventNames, " | "))
	output += unionType

	_, err := fmt.Fprint(g.writer, output)
	return err
}

func (g *TypeScriptTypeGenerator) generateAuditLogEventInterface(action auditlogs.AuditLogAction) string {
	interfaceName := g.generateInterfaceName(action.Name)

	var properties []string

	properties = append(properties, fmt.Sprintf("  action: '%s';", action.Name))
	properties = append(properties, "  occurredAt: Date;")
	properties = append(properties, fmt.Sprintf("  version?: number;"))

	actorProps := []string{
		"    id: string;",
		"    name?: string;",
		"    type: string;",
	}

	if len(action.Schema.Actor.Metadata.Properties) > 0 {
		metadataProps := g.generateMetadataProperties(action.Schema.Actor.Metadata.Properties)
		actorProps = append(actorProps, fmt.Sprintf("    metadata: {\n%s\n    };", metadataProps))
	}

	properties = append(properties, fmt.Sprintf("  actor: {\n%s\n  };", strings.Join(actorProps, "\n")))

	if len(action.Schema.Targets) > 0 {
		var targetTypes []string
		for _, target := range action.Schema.Targets {
			targetProps := []string{
				fmt.Sprintf("    type: '%s';", target.Type),
				"    id: string;",
				"    name?: string;",
			}
			if len(target.Metadata.Properties) > 0 {
				metadataProps := g.generateMetadataProperties(target.Metadata.Properties)
				targetProps = append(targetProps, fmt.Sprintf("    metadata: {\n%s\n    };", metadataProps))
			}

			targetTypes = append(targetTypes, fmt.Sprintf("{\n%s\n  }", strings.Join(targetProps, "\n")))
		}

		properties = append(properties, fmt.Sprintf("  targets: [%s];", strings.Join(targetTypes, " | ")))
	}

	contextProps := []string{
		"    location: string;",
		"    userAgent?: string;",
	}

	properties = append(properties, fmt.Sprintf("  context: {\n%s\n  };", strings.Join(contextProps, "\n")))

	if len(action.Schema.Metadata.Properties) > 0 {
		metadataProps := g.generateMetadataProperties(action.Schema.Metadata.Properties)
		properties = append(properties, fmt.Sprintf("  metadata: {\n%s\n  };", metadataProps))
	}

	return fmt.Sprintf("export interface %s {\n%s\n}\n",
		interfaceName,
		strings.Join(properties, "\n"))
}

func (g *TypeScriptTypeGenerator) generateMetadataProperties(properties map[string]auditlogs.AuditLogActionSchemaMetadataProperty) string {
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
func (g *TypeScriptTypeGenerator) generateInterfaceName(action string) string {
	parts := strings.Split(action, ".")
	var name string
	for _, part := range parts {
		part = strings.ReplaceAll(part, "_", " ")
		words := strings.Fields(part)
		for _, word := range words {
			name += cases.Title(language.English).String(word)
		}
	}

	return fmt.Sprintf("%sAuditLogEvent", name)
}

func (g *TypeScriptTypeGenerator) generateUnionType(typeName string, values []string) error {
	quotedValues := make([]string, len(values))
	for i, v := range values {
		quotedValues[i] = fmt.Sprintf("'%s'", v)
	}

	output := fmt.Sprintf("export type %s = %s;\n\n", typeName, strings.Join(quotedValues, " | "))
	_, err := fmt.Fprint(g.writer, output)

	return err
}
