package main

import (
	"testing"
)

func TestPlanParse(t *testing.T) {
	t.Run("Parses change values", func(t *testing.T) {
		output := `{"@level":"info","@message":"Plan: 8 to add, 7 to change, 8 to destroy.","changes":{"add":8,"change":7,"import":0,"remove":8,"operation":"plan"}}`
		got := parsePlanOutput(output)
		want := TerraformChanges{8, 7, 8}

		if got != want {
			t.Errorf("Expected TerraformChanges{8,7,8}, got %v", got)
		}
	})

	t.Run("Parses error at end", func(t *testing.T) {
		output := `{"@level":"info","@message":"Plan: 8 to add, 7 to change, 8 to destroy.","changes":{"add":8,"change":7,"import":0,"remove":8,"operation":"plan"}}
{"@level":"error","@message":"Error: Unsupported attribute"}`
		got := parsePlanOutput(output)
		want := TerraformChanges{-1, -1, -1}

		if got != want {
			t.Errorf("Expected TerraformChanges{-1,-1,-1}, got %v", got)
		}
	})
}
