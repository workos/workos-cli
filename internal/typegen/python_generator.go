package typegen

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/workos/workos-go/v4/pkg/auditlogs"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type PythonTypeGenerator struct {
	writer io.Writer
}

func NewPythonTypeGenerator(w io.Writer) *PythonTypeGenerator {
	return &PythonTypeGenerator{writer: w}
}

// Import generator
func (g *PythonTypeGenerator) generateImports(resources []string) error {
	if slices.Contains(resources, "all") || slices.Contains(resources, "audit-logs") {
		output := "from typing import Literal, NotRequired, Sequence, TypedDict, Union\n\n"

		_, err := fmt.Fprint(g.writer, output)
		return err
	}

	return nil
}

// Audit Log Generator
func (g *PythonTypeGenerator) generateAuditLogActionsTypes(actions []auditlogs.AuditLogAction) error {
	var interfaces []string
	var eventNames []string

	for _, action := range actions {
		interfaceName := g.generateInterfaceName(action.Name)
		eventNames = append(eventNames, interfaceName)

		interfaceContent := g.generateAuditLogEventInterface(action)
		interfaces = append(interfaces, interfaceContent)
	}

	output := strings.Join(interfaces, "\n")

	unionType := fmt.Sprintf("\nAuditLogEvent = Union[%s]\n",
		strings.Join(eventNames, ", "))
	output += unionType

	_, err := fmt.Fprint(g.writer, output)
	return err
}

func (g *PythonTypeGenerator) generateTypedDictFromMap(name string, properties map[string]string) string {
	generatedCode := fmt.Sprintf("class %s(TypedDict):\n", name)
	propertiesBuffer := new(bytes.Buffer)

	for key, pythonType := range properties {
		fmt.Fprintf(propertiesBuffer, "  %s: %s\n", key, pythonType)
	}

	generatedCode += propertiesBuffer.String() + "\n"
	return generatedCode
}

func (g *PythonTypeGenerator) generateTypedDictFromMetadataProperties(className string, metadataPropertyMap map[string]auditlogs.AuditLogActionSchemaMetadataProperty) string {
	return g.generateTypedDictFromMap(className, g.generateMetadataProperties(metadataPropertyMap))
}

func (g *PythonTypeGenerator) generateAuditLogEventInterface(action auditlogs.AuditLogAction) string {
	interfaceName := g.generateInterfaceName(action.Name)

	var generatedCode string

	// Map of top-level property keys and types
	properties := map[string]string{
		"action":      fmt.Sprintf("Literal['%s']", action.Name),
		"occurred_at": "str",
		"version":     "NotRequired[int]",
	}

	actorProperties := map[string]string{
		"id":   "str",
		"name": "NotRequired[str]",
	}

	if len(action.Schema.Actor.Metadata.Properties) > 0 {
		actorMetadataInterfaceName := interfaceName + "ActorMetadata"
		generatedCode += g.generateTypedDictFromMetadataProperties(actorMetadataInterfaceName, action.Schema.Actor.Metadata.Properties)
		actorProperties["metadata"] = actorMetadataInterfaceName
	}

	// Add actor property
	actorInterfaceName := interfaceName + "Actor"
	generatedCode += g.generateTypedDictFromMap(actorInterfaceName, actorProperties)
	properties["actor"] = actorInterfaceName

	if len(action.Schema.Targets) > 0 {
		var targetTypes []string
		for _, target := range action.Schema.Targets {
			targetProps := map[string]string{
				"type": fmt.Sprintf("Literal['%s']", target.Type),
				"id":   "str",
				"name": "NotRequired[str]",
			}
			if len(target.Metadata.Properties) > 0 {
				targetMetadataInterfaceName := interfaceName + "TargetMetadata"
				generatedCode += g.generateTypedDictFromMetadataProperties(targetMetadataInterfaceName, target.Metadata.Properties)
				targetProps["metadata"] = targetMetadataInterfaceName
			}

			targetInterfaceName := fmt.Sprintf("%s%s%s", interfaceName, cases.Title(language.English).String(target.Type), "Target")
			generatedCode += g.generateTypedDictFromMap(targetInterfaceName, targetProps)
			targetTypes = append(targetTypes, targetInterfaceName)
		}

		if len(targetTypes) > 1 {
			properties["targets"] = fmt.Sprintf("Sequence[Union[%s]]", strings.Join(targetTypes, ", "))
		} else {
			properties["targets"] = fmt.Sprintf("Sequence[%s]", targetTypes[0])
		}
	}

	// Add context property
	contextInterfaceName := interfaceName + "Context"
	generatedCode += g.generateTypedDictFromMap(contextInterfaceName, map[string]string{
		"location":   "str",
		"user_agent": "NotRequired[str]",
	})
	properties["context"] = contextInterfaceName

	// Add top-level metadata property
	if len(action.Schema.Metadata.Properties) > 0 {
		metadataInterfaceName := interfaceName + "Metadata"
		generatedCode += g.generateTypedDictFromMetadataProperties(metadataInterfaceName, action.Schema.Metadata.Properties)
		properties["metadata"] = metadataInterfaceName
	}

	generatedCode += fmt.Sprintf("# Types for %ss\nclass %s(TypedDict):\n", interfaceName, interfaceName)
	propertiesBuffer := new(bytes.Buffer)

	for key, pythonType := range properties {
		fmt.Fprintf(propertiesBuffer, "  %s: %s\n", key, pythonType)
	}

	generatedCode += propertiesBuffer.String()
	return generatedCode
}

func (g *PythonTypeGenerator) generateMetadataProperties(properties map[string]auditlogs.AuditLogActionSchemaMetadataProperty) map[string]string {
	convertedProperties := make(map[string]string)
	for key, prop := range properties {
		if prop.Nullable != nil && *prop.Nullable {
			convertedProperties[key] = fmt.Sprintf("NotRequired[%s]", prop.Type)
		} else {
			convertedProperties[key] = g.convertMetadataType(prop.Type)
		}
	}

	return convertedProperties
}

func (g *PythonTypeGenerator) convertMetadataType(propertyType string) string {
	switch propertyType {
	case "string":
		return "str"
	default:
		return propertyType
	}
}

// Utils
func (g *PythonTypeGenerator) generateInterfaceName(action string) string {
	parts := strings.Split(action, ".")
	var name string
	for _, part := range parts {
		part = strings.ReplaceAll(part, "_", " ")
		words := strings.Fields(part)
		for _, word := range words {
			name += cases.Title(language.English).String(word)
		}
	}

	return name + "AuditLogEvent"
}

func (g *PythonTypeGenerator) generateUnionType(typeName string, values []string) error {
	quotedValues := make([]string, len(values))
	for i, v := range values {
		quotedValues[i] = fmt.Sprintf("'%s'", v)
	}

	output := fmt.Sprintf("%s = Literal[%s]\n\n", typeName, strings.Join(quotedValues, ", "))
	_, err := fmt.Fprint(g.writer, output)

	return err
}
