package main

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

type MockDirEntry struct {
	name  string
	isDir bool
}

func (m MockDirEntry) Name() string {
	return m.name
}

func (m MockDirEntry) IsDir() bool {
	return m.isDir
}

func (m MockDirEntry) Type() fs.FileMode {
	return 1
}

func (m MockDirEntry) Info() (fs.FileInfo, error) {
	return nil, nil
}

func TestFindProjects(t *testing.T) {
	t.Run("project in root", func(t *testing.T) {
		filesystem := fstest.MapFS{
			".terraform/root.go": {},
		}

		want := []fs.DirEntry{MockDirEntry{name: "."}}
		projects, err := findAllTerraformProjects(filesystem)
		if err != nil {
			t.Errorf("Got error %v:", err)
		}
		assertSameDirectories(t, projects, want)
	})

	t.Run("ignores files", func(t *testing.T) {
		filesystem := fstest.MapFS{
			".terraform/terraform.go": {},
			"root.go":                 {},
		}

		want := []fs.DirEntry{MockDirEntry{name: "."}}
		projects, err := findAllTerraformProjects(filesystem)
		if err != nil {
			t.Errorf("Got error %v:", err)
		}
		assertSameDirectories(t, projects, want)
	})

	t.Run("ignores dot directories", func(t *testing.T) {
		filesystem := fstest.MapFS{
			".terraform/terraform.go": {},
			".test/test.go":           {},
		}

		want := []fs.DirEntry{MockDirEntry{name: "."}}
		projects, err := findAllTerraformProjects(filesystem)
		if err != nil {
			t.Errorf("Got error %v:", err)
		}
		assertSameDirectories(t, projects, want)
	})

	t.Run("finds nested project", func(t *testing.T) {
		filesystem := fstest.MapFS{
			".terraform/terraform.go":         {},
			"project/.terraform/terraform.go": {},
		}

		want := []fs.DirEntry{MockDirEntry{name: "."}, MockDirEntry{name: "project"}}
		projects, err := findAllTerraformProjects(filesystem)
		if err != nil {
			t.Errorf("Got error %v:", err)
		}
		assertSameDirectories(t, projects, want)
	})

	t.Run("finds double nested project", func(t *testing.T) {
		filesystem := fstest.MapFS{
			".terraform/terraform.go":             {},
			"project/sub/.terraform/terraform.go": {},
		}

		want := []fs.DirEntry{MockDirEntry{name: "."}, MockDirEntry{name: "sub"}}
		projects, err := findAllTerraformProjects(filesystem)
		if err != nil {
			t.Errorf("Got error %v:", err)
		}
		assertSameDirectories(t, projects, want)
	})
}

func assertSameDirectories(t *testing.T, got []Project, want []fs.DirEntry) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}

	for i := range got {
		if got[i].Name != want[i].Name() {
			t.Errorf("got %v, want %v", got[i], want[i])
		}
	}
}
