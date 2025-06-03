package main

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func TestSpec(t *testing.T) {
	t.Parallel()
	checktest.SpecTest(t, spec)
}

func TestSimpleSuccess(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/simple_success"},
				FilePaths: []string{"simple.proto"},
			},
		},
		Spec: spec,
	}.Run(t)
}

func TestSimpleFailureWithOption(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/simple_failure"},
				FilePaths: []string{"simple.proto"},
			},
			RuleIDs: []string{requiredEntityFieldsRuleID},
			Options: map[string]any{
				requiredEntityFieldsOptionKey: []string{"category"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  requiredEntityFieldsRuleID,
				Message: "field \"updated_at\" is discouraged, use \"last_modified_at\" instead",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   50,
					StartColumn: 4,
					EndLine:     50,
					EndColumn:   45,
				},
			},
			{
				RuleID:  requiredEntityFieldsRuleID,
				Message: "message \"BookCategory\" is missing required fields: [category]",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   53,
					StartColumn: 0,
					EndLine:     60,
					EndColumn:   1,
				},
			},
			{
				RuleID:  requiredEntityFieldsRuleID,
				Message: "field \"last_updated_at\" is discouraged, use \"last_modified_at\" instead",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   59,
					StartColumn: 4,
					EndLine:     59,
					EndColumn:   50,
				},
			},
		},
	}.Run(t)
}

func TestSimpleFailure(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/simple_failure"},
				FilePaths: []string{"simple.proto"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  requiredEntityFieldsRuleID,
				Message: "message \"Book\" is missing required fields: [id account_id created_at]",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   42,
					StartColumn: 0,
					EndLine:     51,
					EndColumn:   1,
				},
			},
			{
				RuleID:  requiredEntityFieldsRuleID,
				Message: "field \"updated_at\" is discouraged, use \"last_modified_at\" instead",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   50,
					StartColumn: 4,
					EndLine:     50,
					EndColumn:   45,
				},
			},
			{
				RuleID:  requiredEntityFieldsRuleID,
				Message: "message \"BookCategory\" is missing required fields: [name]",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   53,
					StartColumn: 0,
					EndLine:     60,
					EndColumn:   1,
				},
			},
			{
				RuleID:  requiredEntityFieldsRuleID,
				Message: "field \"last_updated_at\" is discouraged, use \"last_modified_at\" instead",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   59,
					StartColumn: 4,
					EndLine:     59,
					EndColumn:   50,
				},
			},
			{
				RuleID:  requiredRequestFieldsRuleID,
				Message: "message \"ListBooksRequest\" is missing required fields: [account_id]",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   17,
					StartColumn: 0,
					EndLine:     19,
					EndColumn:   1,
				},
			},
			{
				RuleID:  requiredRequestFieldsRuleID,
				Message: "message \"GetBookRequest\" is missing required fields: [account_id]",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   25,
					StartColumn: 0,
					EndLine:     28,
					EndColumn:   1,
				},
			},
		},
	}.Run(t)
}
