package main

import (
	"testing"
)

func TestPlanParseJSON(t *testing.T) {
	t.Run("Parses change values", func(t *testing.T) {
		output := `{"@level":"info","@message":"Plan: 8 to add, 7 to change, 8 to destroy.","changes":{"add":8,"change":7,"import":0,"remove":8,"operation":"plan"}}`
		got := parsePlanOutputJSON(output)
		want := TerraformChanges{8, 7, 8}

		assertMatchingChanges(t, got, want)
	})

	t.Run("Parses error at end", func(t *testing.T) {
		output := `{"@level":"info","@message":"Plan: 8 to add, 7 to change, 8 to destroy.","changes":{"add":8,"change":7,"import":0,"remove":8,"operation":"plan"}}
{"@level":"error","@message":"Error: Unsupported attribute"}`
		got := parsePlanOutputJSON(output)
		want := TerraformChanges{PlanError.Value(), PlanError.Value(), PlanError.Value()}

		assertMatchingChanges(t, got, want)
	})
}

func TestPlanParse(t *testing.T) {
	t.Run("Parses change values", func(t *testing.T) {
		output := "Plan: 0 to add, 13 to change, 0 to destroy."
		got := parsePlanOutput(output)
		want := TerraformChanges{0, 13, 0}

		assertMatchingChanges(t, got, want)
	})

	t.Run("Parses error at end", func(t *testing.T) {
		output := "Plan: 8 to add, 7 to change, 8 to destroy. Error: Unsupported attribute"
		got := parsePlanOutput(output)
		want := TerraformChanges{PlanError.Value(), PlanError.Value(), PlanError.Value()}

		assertMatchingChanges(t, got, want)
	})

	t.Run("No changes", func(t *testing.T) {
		output := "No changes. Your infrastructure matches the configuration."
		got := parsePlanOutput(output)
		want := TerraformChanges{0, 0, 0}

		assertMatchingChanges(t, got, want)
	})

	t.Run("Outside changes", func(t *testing.T) {
		output := "Objects have changed outside of Terraform"
		got := parsePlanOutput(output)
		want := TerraformChanges{DriftError.Value(), DriftError.Value(), DriftError.Value()}

		assertMatchingChanges(t, got, want)
	})
}

func TestRegexMatch(t *testing.T) {
	t.Run("Parses change values", func(t *testing.T) {
		output := "Plan: 8 to add, 7 to change, 8 to destroy."
		got, err := regexMatchChanges(output)
		want := TerraformChanges{8, 7, 8}

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		assertMatchingChanges(t, got, want)
	})
}

func assertMatchingChanges(t *testing.T, got, want TerraformChanges) {
	t.Helper()
	if got != want {
		t.Errorf("Expected TerraformChanges{%d,%d,%d}, got %v", want.Add, want.Change, want.Destroy, got)
	}
}
