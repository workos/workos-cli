package typegen

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/workos/workos-go/v4/pkg/auditlogs"
	"github.com/workos/workos-go/v4/pkg/permissions"
	"github.com/workos/workos-go/v4/pkg/roles"
)

type TypeGenerator interface {
	GeneratePermissionsType() error
	GenerateRolesType() error
	GenerateAuditLogsTypes() error
}

type LanguageTypeGenerator interface {
	generateImports(resources []string) error
	generateUnionType(typeName string, values []string) error
	generateInterfaceName(resourceName string) string
	generateAuditLogActionsTypes(actions []auditlogs.AuditLogAction) error
}

// Language Import Generator
func GenerateImports(languageGenerator LanguageTypeGenerator, resources []string) error {
	return languageGenerator.generateImports(resources)
}

// Permission Generator
func GeneratePermissionsType(languageGenerator LanguageTypeGenerator) error {
	// TODO: Figure out why we need to set the API key here
	permissions.SetAPIKey(os.Getenv("WORKOS_API_KEY"))

	// TODO: Need to support pagination
	permissionsList, err := permissions.ListPermissions(
		context.Background(),
		permissions.ListPermissionsOpts{},
	)
	if err != nil {
		return errors.Wrap(err, "error fetching permissions")
	}

	slugs := make([]string, len(permissionsList.Data))
	for i, permission := range permissionsList.Data {
		slugs[i] = permission.Slug
	}

	return languageGenerator.generateUnionType("WorkOSPermission", slugs)
}

// Role Generator
func GenerateRolesType(languageGenerator LanguageTypeGenerator) error {
	// TODO: Figure out why we need to set the API key here
	roles.SetAPIKey(os.Getenv("WORKOS_API_KEY"))

	rolesList, err := roles.ListRoles(
		context.Background(),
		roles.ListRolesOpts{},
	)
	if err != nil {
		return errors.Wrap(err, "error fetching roles")
	}

	slugs := make([]string, len(rolesList.Data))
	for i, role := range rolesList.Data {
		slugs[i] = role.Slug
	}

	return languageGenerator.generateUnionType("WorkOSRole", slugs)
}

// Audit Log Generator
func GenerateAuditLogsTypes(languageGenerator LanguageTypeGenerator) error {
	auditlogs.SetAPIKey(os.Getenv("WORKOS_API_KEY"))

	// TODO: Need to support pagination
	auditLogsList, err := auditlogs.ListActions(
		context.Background(),
		auditlogs.ListActionsOpts{},
	)
	if err != nil {
		return errors.Wrap(err, "error fetching audit logs actions")
	}

	return languageGenerator.generateAuditLogActionsTypes(auditLogsList.Data)
}
