// Package main implements a plugin that checks that:
// - entity-related messages (e.g: Cluster) define a known set of common fields
// for the Qdrant Cloud API. Default values: id, name, account_id, created_at
// - Request messages (e.g: ListClustersRequest) define a known set of common fields
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
	"fmt"
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

// FieldValidator validates a single field.
// Returns an error message and false if validation fails.
type FieldValidator func(field protoreflect.FieldDescriptor) *ValidationError

// MessageValidator validates a message as a whole, based on the set of fields present in the message.
// Returns an error message and false if validation fails.
type MessageValidator func(message protoreflect.MessageDescriptor, messageFields map[string]bool) *ValidationError

// ValidationError represents a linting error and includes the error message and
// the descriptor where the linting issue was found.
type ValidationError struct {
	Message    string
	Descriptor protoreflect.Descriptor
}

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
	preferredEntityFieldNames           = map[string]string{
		"updated_at":            "last_modified_at",
		"last_updated_at":       "last_modified_at",
		"cloud_provider":        "cloud_provider_id",
		"cloud_provider_region": "cloud_provider_region_id",
		"cloud_region":          "cloud_provider_region_id",
		"cloud_region_id":       "cloud_provider_region_id",
	}
)

func main() {
	check.Main(spec)
}

// checkEntityFields validates all entity-related messages in a file descriptor.
// It applies:
// - Field-level validators (e.g. preferred naming).
// - Message-level validators (e.g. required fields).
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
		errors := validateMessage(
			msg,
			[]FieldValidator{preferredFieldNamesValidator(preferredEntityFieldNames)},
			[]MessageValidator{missingFieldsValidator(requiredFields)},
		)

		for _, err := range errors {
			responseWriter.AddAnnotation(check.WithMessage(err.Message), check.WithDescriptor(err.Descriptor))
		}
	}

	return nil
}

// checkRequestFields validates messages that end with "Request" and match a known
// CRUD pattern (e.g., ListClustersRequest). It ensures these messages include required fields.
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
	errors := validateMessage(
		messageDescriptor, []FieldValidator{}, []MessageValidator{missingFieldsValidator(requiredFields)},
	)
	for _, err := range errors {
		responseWriter.AddAnnotation(check.WithMessage(err.Message), check.WithDescriptor(err.Descriptor))
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
// e.g: [ListBooks, GetBook] -> {Book}.
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

// inferEntityFromMethodName extracts the entity name by stripping CRUD prefixes.
func inferEntityFromMethodName(methodName string) string {
	p := pluralize.NewClient()
	for _, prefix := range crudMethodPrefixes {
		if strings.HasPrefix(methodName, prefix) {
			return p.Singular(strings.TrimPrefix(methodName, prefix))
		}
	}
	return ""
}

// validateMessage runs a set of field-level and message-level validators
// against a protobuf message descriptor.
//
// Field-level validators are executed for each individual field in the message,
// allowing checks like discouraged field names or naming conventions.
//
// Message-level validators are run once per message, and have access to the
// full set of field names, enabling checks like required field presence.
func validateMessage(msg protoreflect.MessageDescriptor, fieldValidators []FieldValidator, messageValidators []MessageValidator) []ValidationError {
	// missingFields := []string{}
	existingFields := make(map[string]bool)
	fields := msg.Fields()
	errors := []ValidationError{}

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldName := string(field.Name())
		existingFields[string(fieldName)] = true

		for _, validator := range fieldValidators {
			if err := validator(field); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	for _, validator := range messageValidators {
		if err := validator(msg, existingFields); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// preferredFieldNamesValidator returns a FieldValidator that checks
// if a given field name is discouraged and suggests the preferred one.
func preferredFieldNamesValidator(preferredFieldNames map[string]string) FieldValidator {
	return func(field protoreflect.FieldDescriptor) *ValidationError {
		fieldName := string(field.Name())
		if suggestion, ok := preferredFieldNames[fieldName]; ok && suggestion != fieldName {
			return &ValidationError{
				Message:    fmt.Sprintf("field %q is discouraged, use %q instead", fieldName, suggestion),
				Descriptor: field,
			}
		}
		return nil
	}
}

// missingFieldsValidator returns a MessageValidator that ensures a message
// contains all of the specified required fields.
func missingFieldsValidator(requiredFields []string) MessageValidator {
	return func(message protoreflect.MessageDescriptor, messageFields map[string]bool) *ValidationError {
		messageName := string(message.Name())
		missingFields := []string{}
		for _, requiredField := range requiredFields {
			if !messageFields[requiredField] {
				missingFields = append(missingFields, requiredField)
			}
		}
		if len(missingFields) > 0 {
			return &ValidationError{
				Message:    fmt.Sprintf("message %q is missing required fields: %v", messageName, missingFields),
				Descriptor: message,
			}
		}
		return nil
	}
}
