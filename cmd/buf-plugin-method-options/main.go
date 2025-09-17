// Package main implements a plugin that checks that all rpc methods set the
// required options. The list of options is configurable.
// The default value is:
// - "qdrant.cloud.common.v1.permissions"
// - "google.api.http"
//
// To use this plugin:
//
//	# buf.yaml
//	version: v2
//	lint:
//	  use:
//	   - STANDARD # omit if you do not want to use the rules builtin to buf
//	   - QDRANT_CLOUD_METHOD_OPTIONS
//	plugins:
//	  - plugin: buf-plugin-method-options
//	    # Uncomment in case you need to configure the list of method options to validate.
//	    # options:
//	    #  required_method_options:
//	    #    - "qdrant.cloud.common.v1.permissions"
package main

import (
	"context"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/info"
	"buf.build/go/bufplugin/option"
	googleann "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"

	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
)

const (
	// methodOptionsRuleID is the Rule ID of the methodOptions rule.
	methodOptionsRuleID = "QDRANT_CLOUD_METHOD_OPTIONS"
	// methodOptionsOptionKey is the option key to override the default list of required options.
	methodOptionsOptionKey = "required_method_options"
)

var (
	methodOptionsRuleSpec = &check.RuleSpec{
		ID:      methodOptionsRuleID,
		Default: true,
		Purpose: `Checks that all rpc methods define a set of required options.`,
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewMethodRuleHandler(checkMethodOptions, checkutil.WithoutImports()),
	}
	spec = &check.Spec{
		Rules: []*check.RuleSpec{
			methodOptionsRuleSpec,
		},
		Info: &info.Spec{
			Documentation: `A plugin that checks that all rpc methods define a set of required options.`,
			SPDXLicenseID: "",
			LicenseURL:    "",
		},
	}
	permissionsOption            = commonv1.E_Permissions
	restHTTPOption               = googleann.E_Http
	requiresAuthenticationOption = commonv1.E_RequiresAuthentication
	accountIdExpressionOption    = commonv1.E_AccountIdExpression

	extensionRegistry = map[string]*protoimpl.ExtensionInfo{
		string(permissionsOption.TypeDescriptor().Descriptor().FullName()): permissionsOption,
		string(restHTTPOption.TypeDescriptor().Descriptor().FullName()):    restHTTPOption,
	}
	requiredMethodOptionExtensions = []string{
		string(permissionsOption.TypeDescriptor().Descriptor().FullName()),
		string(restHTTPOption.TypeDescriptor().Descriptor().FullName()),
	}
)

func main() {
	check.Main(spec)
}

func checkMethodOptions(ctx context.Context, responseWriter check.ResponseWriter, request check.Request, methodDescriptor protoreflect.MethodDescriptor) error {
	requiredOptions := requiredMethodOptionExtensions
	optionValue, err := option.GetStringSliceValue(request.Options(), methodOptionsOptionKey)
	if err != nil {
		return err
	}
	if len(optionValue) > 0 {
		requiredOptions = optionValue
	}

	options := methodDescriptor.Options()

	for _, extensionKey := range requiredOptions {
		extension, found := extensionRegistry[extensionKey]
		if !found {
			responseWriter.AddAnnotation(
				check.WithMessagef("extension key %q does not exist", extensionKey),
			)
			return nil
		}
		if !proto.HasExtension(options, extension) {
			// special case for "qdrant.cloud.common.v1.permissions": in case
			// there is "qdrant.cloud.common.v1.requires_authentication" set to
			// false, setting permissions isn't needed.
			if extensionKey == "qdrant.cloud.common.v1.permissions" && proto.HasExtension(options, requiresAuthenticationOption) {
				val := proto.GetExtension(options, requiresAuthenticationOption).(bool)
				if !val {
					// requires_authentication is false, we skip it.
					break
				}
			}
			responseWriter.AddAnnotation(
				check.WithMessagef("Method %q does not define the %q option", methodDescriptor.FullName(), extension.TypeDescriptor().FullName()),
				check.WithDescriptor(methodDescriptor),
			)
		}
	}

	// Check for permissions + account_id_expression conflict
	if proto.HasExtension(options, permissionsOption) && proto.HasExtension(options, accountIdExpressionOption) {
		permissionsExpression := proto.GetExtension(options, permissionsOption).([]string)
		accountIdExpression := proto.GetExtension(options, accountIdExpressionOption).(string)

		var permissions []string
		for _, perm := range permissionsExpression {
			if perm != "" {
				permissions = append(permissions, perm)
			}
		}

		// If there are permissions but account_id_expression is empty,
		// this is invalid because permissions are checked in the scope of the account
		if len(permissions) > 0 && accountIdExpression == "" {
			responseWriter.AddAnnotation(
				check.WithMessagef("Method %q has permissions set but account_id_expression is empty. Methods with permissions require a non-empty account_id_expression since permissions are checked in the scope of the account", methodDescriptor.FullName()),
				check.WithDescriptor(methodDescriptor),
			)
		}
	}

	return nil
}
