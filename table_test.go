package main

import (
	"testing"
)

func TestMatchHighlightedProject(t *testing.T) {
	projects := []Project{
		{
			Name: "pName",
			Path: "pPath",
		},
	}

	t.Run("Match", func(t *testing.T) {
		project := matchHighlightedProject("pPath", &projects)
		if project == nil {
			t.Error("Expected to find a project")
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		project := matchHighlightedProject("noPath", &projects)
		if project != nil {
			t.Error("Expected to not find a project")
		}
	})
}
