// Package main implements a plugin that checks for breaking changes in permissions.
// The plugin detects when method permissions change, which could break existing
// client code that depends on certain access levels.
//
// The plugin handles both AND and OR permission logic via the requires_all_permissions option:
// - requires_all_permissions=true (default): ALL permissions must be met (AND logic)
// - requires_all_permissions=false: ANY ONE permission must be met (OR logic)
//
// Breaking changes detected:
// - Adding permissions to methods that previously had none (restricts access)
// - Removing all permissions from methods (changes access model)
// - Changing requires_all_permissions from false to true (OR to AND, more restrictive)
// - For AND permissions (requires_all_permissions=true): ANY change to permissions
// - For OR permissions (requires_all_permissions=false): REMOVING permissions
//
// Non-breaking changes (not reported):
// - New methods with permissions (handled automatically by buf framework)
// - Changing requires_all_permissions from true to false (AND to OR, more permissive)
// - For OR permissions (requires_all_permissions=false): ADDING permissions
//
// To use this plugin:
//
//	# buf.yaml
//	version: v2
//	breaking:
//	  use:
//	   - QDRANT_CLOUD_PERMISSIONS_BREAKING
//	plugins:
//	  - plugin: buf-plugin-permissions-breaking
package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/info"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
)

const (
	permissionsBreakingRuleID = "QDRANT_CLOUD_PERMISSIONS_BREAKING"
)

// PermissionConfig holds the permission configuration for a method.
type PermissionConfig struct {
	Permissions []string
	RequiresAll bool // true = AND (default), false = OR
}

var (
	permissionsBreakingRuleSpec = &check.RuleSpec{
		ID:      permissionsBreakingRuleID,
		Default: true,
		Purpose: `Checks for breaking changes in method permissions.`,
		Type:    check.RuleTypeBreaking,
		Handler: checkutil.NewMethodPairRuleHandler(checkPermissionsBreaking, checkutil.WithoutImports()),
	}
	spec = &check.Spec{
		Rules: []*check.RuleSpec{
			permissionsBreakingRuleSpec,
		},
		Info: &info.Spec{
			Documentation: `A plugin that checks for breaking changes in method permissions.`,
			SPDXLicenseID: "",
			LicenseURL:    "",
		},
	}
	permissionsOption            = commonv1.E_Permissions
	requiresAllPermissionsOption = commonv1.E_RequiresAllPermissions
)

func main() {
	check.Main(spec)
}

func checkPermissionsBreaking(ctx context.Context, responseWriter check.ResponseWriter, request check.Request, methodDescriptor, againstMethodDescriptor protoreflect.MethodDescriptor) error {
	againstConfig := getMethodPermissionConfig(againstMethodDescriptor)
	currentConfig := getMethodPermissionConfig(methodDescriptor)

	// Check for breaking changes based on permission logic
	if isBreakingChange(againstConfig, currentConfig) {
		var message string
		if len(currentConfig.Permissions) == 0 {
			message = fmt.Sprintf("Method %q had permissions %v but now has no permissions, this is a breaking change",
				methodDescriptor.FullName(), againstConfig.Permissions)
		} else if len(againstConfig.Permissions) == 0 {
			message = fmt.Sprintf("Method %q had no permissions but now requires permissions %v, this is a breaking change",
				methodDescriptor.FullName(), currentConfig.Permissions)
		} else {
			requiresAllChanged := againstConfig.RequiresAll != currentConfig.RequiresAll
			if requiresAllChanged {
				message = fmt.Sprintf("Method %q permissions logic changed from requires_all=%t to requires_all=%t with permissions %v to %v, this is a breaking change",
					methodDescriptor.FullName(), againstConfig.RequiresAll, currentConfig.RequiresAll, againstConfig.Permissions, currentConfig.Permissions)
			} else {
				message = fmt.Sprintf("Method %q permissions changed from %v to %v (requires_all=%t), this is a breaking change",
					methodDescriptor.FullName(), againstConfig.Permissions, currentConfig.Permissions, currentConfig.RequiresAll)
			}
		}
		responseWriter.AddAnnotation(
			check.WithMessage(message),
			check.WithDescriptor(methodDescriptor),
		)
	}

	return nil
}

// getMethodPermissionConfig extracts the permission configuration from a method descriptor.
func getMethodPermissionConfig(methodDescriptor protoreflect.MethodDescriptor) PermissionConfig {
	options := methodDescriptor.Options()

	// Extract permissions
	var permissions []string
	if proto.HasExtension(options, permissionsOption) {
		permissionsRaw := proto.GetExtension(options, permissionsOption)
		if permissionsSlice, ok := permissionsRaw.([]string); ok {
			// Filter out empty permissions and sort for consistent comparison
			for _, perm := range permissionsSlice {
				if strings.TrimSpace(perm) != "" {
					permissions = append(permissions, strings.TrimSpace(perm))
				}
			}
			sort.Strings(permissions)
		}
	}

	// Extract requires_all_permissions (defaults to true)
	requiresAll := true // Default to AND behavior
	if proto.HasExtension(options, requiresAllPermissionsOption) {
		if val, ok := proto.GetExtension(options, requiresAllPermissionsOption).(bool); ok {
			requiresAll = val
		}
	}

	return PermissionConfig{
		Permissions: permissions,
		RequiresAll: requiresAll,
	}
}

// isBreakingChange determines if a permission configuration change is breaking.
func isBreakingChange(against, current PermissionConfig) bool {
	// If both configs are identical, no breaking change
	if configsEqual(against, current) {
		return false
	}

	// If requires_all_permissions logic changed:
	// - true -> false (AND to OR): non-breaking (more permissive)
	// - false -> true (OR to AND): breaking (more restrictive)
	if against.RequiresAll != current.RequiresAll {
		if against.RequiresAll && !current.RequiresAll {
			// Changed from AND to OR - non-breaking (more permissive)
			return false
		} else {
			// Changed from OR to AND - breaking (more restrictive)
			return true
		}
	}

	// Handle the case where permissions are added to a method that had none
	if len(against.Permissions) == 0 && len(current.Permissions) > 0 {
		return true // Adding permissions to a previously unrestricted method is breaking
	}

	// Handle the case where permissions are removed completely
	if len(against.Permissions) > 0 && len(current.Permissions) == 0 {
		return true // Removing all permissions changes the access model
	}

	// For methods that had permissions before and still have permissions
	if len(against.Permissions) > 0 && len(current.Permissions) > 0 {
		if against.RequiresAll {
			// AND logic: ANY change is breaking (both adding and removing permissions)
			return !permissionsEqual(against.Permissions, current.Permissions)
		} else {
			// OR logic: Only removing permissions is breaking, adding is non-breaking
			return hasRemovedPermissions(against.Permissions, current.Permissions)
		}
	}

	return false
}

// configsEqual checks if two permission configurations are identical.
func configsEqual(a, b PermissionConfig) bool {
	return a.RequiresAll == b.RequiresAll && permissionsEqual(a.Permissions, b.Permissions)
}

// hasRemovedPermissions checks if any permissions were removed (for OR logic).
func hasRemovedPermissions(previous, current []string) bool {
	currentSet := make(map[string]bool)
	for _, perm := range current {
		currentSet[perm] = true
	}

	for _, perm := range previous {
		if !currentSet[perm] {
			return true // Found a permission that was removed
		}
	}
	return false
}

func permissionsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
