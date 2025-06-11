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
				RuleID:  methodOptionsRuleID,
				Message: "Method \"simple.GreeterService.HelloWorld\" does not define the \"google.api.http\" option",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   9,
					StartColumn: 4,
					EndLine:     12,
					EndColumn:   5,
				},
			},
			{
				RuleID:  methodOptionsRuleID,
				Message: "Method \"simple.GreeterService.HelloWorld\" does not define the \"qdrant.cloud.common.v1.permissions\" option",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   9,
					StartColumn: 4,
					EndLine:     12,
					EndColumn:   5,
				},
			},
			{
				RuleID:  methodOptionsRuleID,
				Message: "Method \"simple.GreeterService.ClosedGoodbye\" does not define the \"google.api.http\" option",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   14,
					StartColumn: 4,
					EndLine:     18,
					EndColumn:   5,
				},
			},
			{
				RuleID:  methodOptionsRuleID,
				Message: "Method \"simple.GreeterService.ClosedGoodbye\" does not define the \"qdrant.cloud.common.v1.permissions\" option",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   14,
					StartColumn: 4,
					EndLine:     18,
					EndColumn:   5,
				},
			},
		},
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
			Options: map[string]any{
				methodOptionsOptionKey: []string{"qdrant.cloud.common.v1.permissions", "unknown.extension"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  methodOptionsRuleID,
				Message: "extension key \"unknown.extension\" does not exist",
			},
			{
				RuleID:  methodOptionsRuleID,
				Message: "extension key \"unknown.extension\" does not exist",
			},
			{
				RuleID:  methodOptionsRuleID,
				Message: "Method \"simple.GreeterService.HelloWorld\" does not define the \"qdrant.cloud.common.v1.permissions\" option",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   9,
					StartColumn: 4,
					EndLine:     12,
					EndColumn:   5,
				},
			},
			{
				RuleID:  methodOptionsRuleID,
				Message: "Method \"simple.GreeterService.ClosedGoodbye\" does not define the \"qdrant.cloud.common.v1.permissions\" option",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "simple.proto",
					StartLine:   14,
					StartColumn: 4,
					EndLine:     18,
					EndColumn:   5,
				},
			},
		},
	}.Run(t)

}

func TestSimpleFailureWithOptionWrongKey(t *testing.T) {
	t.Parallel()
	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/simple_failure"},
				FilePaths: []string{"simple.proto"},
			},
			Options: map[string]any{
				methodOptionsOptionKey: []string{"unknown.extension"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  methodOptionsRuleID,
				Message: "extension key \"unknown.extension\" does not exist",
			},
			{
				RuleID:  methodOptionsRuleID,
				Message: "extension key \"unknown.extension\" does not exist",
			},
		},
	}.Run(t)

}
