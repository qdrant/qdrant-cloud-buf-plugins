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
	extensionRegistry = map[string]*protoimpl.ExtensionInfo{
		"qdrant.cloud.common.v1.permissions": commonv1.E_Permissions,
		"google.api.http":                    googleann.E_Http,
	}
	requiredMethodOptionExtensions = []string{"qdrant.cloud.common.v1.permissions", "google.api.http"}
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
			responseWriter.AddAnnotation(
				check.WithMessagef("Method %q does not define the %q option", methodDescriptor.FullName(), extension.TypeDescriptor().FullName()),
				check.WithDescriptor(methodDescriptor),
			)
		}
	}

	return nil
}
