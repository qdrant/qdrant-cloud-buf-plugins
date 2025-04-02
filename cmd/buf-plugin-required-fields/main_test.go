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
				Message: "\"BookCategory\" is missing required fields: [category]",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   51,
					StartColumn: 0,
					EndLine:     56,
					EndColumn:   1,
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
				Message: "\"Book\" is missing required fields: [id account_id created_at]",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   42,
					StartColumn: 0,
					EndLine:     49,
					EndColumn:   1,
				},
			},
			{
				RuleID:  requiredEntityFieldsRuleID,
				Message: "\"BookCategory\" is missing required fields: [name]",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   51,
					StartColumn: 0,
					EndLine:     56,
					EndColumn:   1,
				},
			},
			{
				RuleID:  requiredRequestFieldsRuleID,
				Message: "\"ListBooksRequest\" is missing required fields: [account_id]",
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
				Message: "\"GetBookRequest\" is missing required fields: [account_id]",
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
