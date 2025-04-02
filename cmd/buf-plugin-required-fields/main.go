// Package main implements a plugin that checks that:
// - entity-related messages (e.g: Cluster) define a known set of common fields
// for the Qdrant Cloud API. Default values: id, name, account_id, created_at
// - Request messages (e.g: ListClusters) define a known set of common fields
// for the Qdrant Cloud API. Default values: account_id
//
// To use this plugin:
//
//	# buf.yaml
//	version: v2
//	lint:
//	  use:
//	   - STANDARD # omit if you do not want to use the rules builtin to buf
//	   - QDRANT_CLOUD_REQUIRED_ENTITY_FIELDS
//	   - QDRANT_CLOUD_REQUIRED_REQUEST_FIELDS
//	plugins:
//	  - plugin: buf-plugin-required-fields
package main

import (
	"context"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	"buf.build/go/bufplugin/descriptor"
	"buf.build/go/bufplugin/info"
	"buf.build/go/bufplugin/option"
	pluralize "github.com/gertd/go-pluralize"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	requiredEntityFieldsRuleID     = "QDRANT_CLOUD_REQUIRED_ENTITY_FIELDS"
	requiredEntityFieldsOptionKey  = "required_entity_fields"
	requiredRequestFieldsRuleID    = "QDRANT_CLOUD_REQUIRED_REQUEST_FIELDS"
	requiredRequestFieldsOptionKey = "required_request_fields"
)

var (
	requiredEntityFieldsRuleSpec = &check.RuleSpec{
		ID:      requiredEntityFieldsRuleID,
		Default: true,
		Purpose: `Checks that all entity-related messages (e.g: Cluster) define a known set of fields for the Qdrant Cloud API.`,
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewFileRuleHandler(checkEntityFields, checkutil.WithoutImports()),
	}
	requiredRequestFieldsRuleSpec = &check.RuleSpec{
		ID:      requiredRequestFieldsRuleID,
		Default: true,
		Purpose: `Checks that all request methods (e.g: ListClustersRequest) define a known set of fields for the Qdrant Cloud API.`,
		Type:    check.RuleTypeLint,
		Handler: checkutil.NewMessageRuleHandler(checkRequestFields, checkutil.WithoutImports()),
	}
	spec = &check.Spec{
		Rules: []*check.RuleSpec{
			requiredEntityFieldsRuleSpec,
			requiredRequestFieldsRuleSpec,
		},
		Info: &info.Spec{
			Documentation: `A plugin that checks that entity-related messages define a known set of fields for the Qdrant Cloud API.`,
			SPDXLicenseID: "",
			LicenseURL:    "",
		},
	}

	crudMethodPrefixes                  = []string{"List", "Get", "Delete", "Update", "Create"}
	crudMethodWithoutFullEntityPrefixes = []string{"List", "Get", "Delete"}
	defaultRequiredFields               = []string{"id", "name", "account_id", "created_at"}
	defaultRequiredRequestFields        = []string{"account_id"}
)

func main() {
	check.Main(spec)
}

func checkEntityFields(ctx context.Context, responseWriter check.ResponseWriter, request check.Request, fileDescriptor descriptor.FileDescriptor) error {
	requiredFields, err := getRequiredEntityFields(request)
	if err != nil {
		return err
	}

	for entityName := range extractEntityNames(fileDescriptor) {
		msg := fileDescriptor.ProtoreflectFileDescriptor().Messages().ByName(protoreflect.Name(entityName))
		if msg == nil {
			continue
		}
		missingFields := findMissingFields(msg, requiredFields)
		if len(missingFields) > 0 {
			responseWriter.AddAnnotation(
				check.WithMessagef("%q is missing required fields: %v", entityName, missingFields),
				check.WithDescriptor(msg),
			)
		}
	}

	return nil
}

func checkRequestFields(ctx context.Context, responseWriter check.ResponseWriter, request check.Request, messageDescriptor protoreflect.MessageDescriptor) error {
	msgName := string(messageDescriptor.Name())
	if !strings.HasSuffix(msgName, "Request") {
		return nil
	}
	var requiredFields []string
	// For Create/Update methods it would be useful to check for the
	// `{entity}_id` field. We could add it later as an improvement.
	for _, prefix := range crudMethodWithoutFullEntityPrefixes {
		if strings.HasPrefix(msgName, prefix) {
			requiredFields = defaultRequiredRequestFields
		}
	}
	missingFields := findMissingFields(messageDescriptor, requiredFields)
	if len(missingFields) > 0 {
		responseWriter.AddAnnotation(
			check.WithMessagef("%q is missing required fields: %v", msgName, missingFields),
			check.WithDescriptor(messageDescriptor),
		)
	}

	return nil
}

// getRequiredEntityFields returns a list of required fields for a entity
// message. It gets the values either from a plugin option or from the default
// values.
func getRequiredEntityFields(request check.Request) ([]string, error) {
	requiredFieldsOptionValue, err := option.GetStringSliceValue(request.Options(), requiredEntityFieldsOptionKey)
	if err != nil {
		return nil, err
	}
	if len(requiredFieldsOptionValue) > 0 {
		return requiredFieldsOptionValue, nil
	}
	return defaultRequiredFields, nil
}

// extractEntityNames returns a set of entity names inferred from the name of
// the service methods.
// e.g: [ListBooks, GetBook] -> {Book}
func extractEntityNames(fileDescriptor descriptor.FileDescriptor) map[string]struct{} {
	entityNames := make(map[string]struct{})
	services := fileDescriptor.FileDescriptorProto().GetService()
	for _, svc := range services {
		for _, method := range svc.Method {
			entityName := inferEntityFromMethodName(method.GetName())
			if entityName != "" {
				entityNames[entityName] = struct{}{}
			}
		}
	}
	return entityNames
}

// inferEntityFromMethodName extracts the entity name by stripping CRUD prefixes
func inferEntityFromMethodName(methodName string) string {
	p := pluralize.NewClient()
	for _, prefix := range crudMethodPrefixes {
		if strings.HasPrefix(methodName, prefix) {
			return p.Singular(strings.TrimPrefix(methodName, prefix))
		}
	}
	return ""
}

// findMissingFields checks if a message contains all required fields.
func findMissingFields(msg protoreflect.MessageDescriptor, requiredFields []string) []string {
	missingFields := []string{}
	fieldMap := make(map[string]bool)
	fields := msg.Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldMap[string(field.Name())] = true
	}

	for _, requiredField := range requiredFields {
		if !fieldMap[requiredField] {
			missingFields = append(missingFields, requiredField)
		}
	}
	return missingFields
}
