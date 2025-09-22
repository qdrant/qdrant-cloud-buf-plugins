package main

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func TestSpec(t *testing.T) {
	t.Parallel()
	checktest.SpecTest(t, spec)
}

func TestBreakingChange(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/breaking_change/current"},
				FilePaths: []string{"service.proto"},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/breaking_change/previous"},
				FilePaths: []string{"service.proto"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  permissionsBreakingRuleID,
				Message: "Method \"test.TestService.TestMethod\" permissions changed from [read:test write:test] to [read:test] (requires_all=true), this is a breaking change",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "service.proto",
					StartLine:   9,
					StartColumn: 2,
					EndLine:     11,
					EndColumn:   3,
				},
			},
		},
	}.Run(t)
}

func TestNewMethodNonBreaking(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/new_method/current"},
				FilePaths: []string{"service.proto"},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/new_method/previous"},
				FilePaths: []string{"service.proto"},
			},
		},
		Spec: spec,
		// No expected annotations - new methods with permissions should not be breaking
	}.Run(t)
}

func TestAddPermissionsBreaking(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/add_permissions/current"},
				FilePaths: []string{"service.proto"},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/add_permissions/previous"},
				FilePaths: []string{"service.proto"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  permissionsBreakingRuleID,
				Message: "Method \"test.TestService.PublicMethod\" had no permissions but now requires permissions [read:restricted], this is a breaking change",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "service.proto",
					StartLine:   9,
					StartColumn: 2,
					EndLine:     11,
					EndColumn:   3,
				},
			},
		},
	}.Run(t)
}

func TestOrPermissionsAddNonBreaking(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/or_permissions_add_non_breaking/current"},
				FilePaths: []string{"service.proto"},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/or_permissions_add_non_breaking/previous"},
				FilePaths: []string{"service.proto"},
			},
		},
		Spec: spec,
		// No expected annotations - adding permissions with OR logic is non-breaking
	}.Run(t)
}

func TestOrPermissionsRemoveBreaking(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/or_permissions_remove_breaking/current"},
				FilePaths: []string{"service.proto"},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/or_permissions_remove_breaking/previous"},
				FilePaths: []string{"service.proto"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  permissionsBreakingRuleID,
				Message: "Method \"test.TestService.FlexibleMethod\" permissions changed from [read:advanced read:basic] to [read:basic] (requires_all=false), this is a breaking change",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "service.proto",
					StartLine:   9,
					StartColumn: 2,
					EndLine:     12,
					EndColumn:   3,
				},
			},
		},
	}.Run(t)
}

func TestAndToOrNonBreaking(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/and_to_or_non_breaking/current"},
				FilePaths: []string{"service.proto"},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/and_to_or_non_breaking/previous"},
				FilePaths: []string{"service.proto"},
			},
		},
		Spec: spec,
		// No expected annotations - changing from AND to OR is non-breaking (more permissive)
	}.Run(t)
}

func TestOrToAndBreaking(t *testing.T) {
	t.Parallel()

	checktest.CheckTest{
		Request: &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/or_to_and_breaking/current"},
				FilePaths: []string{"service.proto"},
			},
			AgainstFiles: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/or_to_and_breaking/previous"},
				FilePaths: []string{"service.proto"},
			},
		},
		Spec: spec,
		ExpectedAnnotations: []checktest.ExpectedAnnotation{
			{
				RuleID:  permissionsBreakingRuleID,
				Message: "Method \"test.TestService.MyMethod\" permissions logic changed from requires_all=false to requires_all=true with permissions [read:data write:data] to [read:data write:data], this is a breaking change",
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "service.proto",
					StartLine:   9,
					StartColumn: 2,
					EndLine:     13,
					EndColumn:   3,
				},
			},
		},
	}.Run(t)
}
