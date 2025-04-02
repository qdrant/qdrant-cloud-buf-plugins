// Package main implements a plugin that checks that all rpc methods set the
// required options (permissions, http).
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
package main

import (
	"context"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/info"
	googleann "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"

	commonv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/common/v1"
)

const (
	methodOptionsRuleID = "QDRANT_CLOUD_METHOD_OPTIONS"
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
	requiredMethodOptionExtensions = []*protoimpl.ExtensionInfo{
		commonv1.E_Permissions,
		googleann.E_Http,
	}
)

func main() {
	check.Main(spec)
}

func checkMethodOptions(ctx context.Context, responseWriter check.ResponseWriter, request check.Request, methodDescriptor protoreflect.MethodDescriptor) error {
	options := methodDescriptor.Options()

	for _, extension := range requiredMethodOptionExtensions {
		if !proto.HasExtension(options, extension) {
			responseWriter.AddAnnotation(
				check.WithMessagef("Method %q does not define the %q option", methodDescriptor.FullName(), extension.TypeDescriptor().FullName()),
				check.WithDescriptor(methodDescriptor),
			)
		}
	}

	return nil
}
