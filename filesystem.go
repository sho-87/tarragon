package main

import (
	"fmt"
	"io/fs"
	"os"
)

const TerraformDir = ".terraform"

func isTerraformProject(filesystem fs.FS, d fs.DirEntry, path string) bool {
	if !d.IsDir() {
		return false
	}

	// fmt.Println("Checking", path)
	dir, err := fs.ReadDir(filesystem, path)
	if err != nil {
		fmt.Println("Error reading directory", path)
		panic(err)
	}

	for _, currentDir := range dir {
		if currentDir.IsDir() && currentDir.Name() == TerraformDir {
			return true
		}
	}
	return false
}

func findAllTerraformProjects(filesystem fs.FS) []fs.DirEntry {
	projects := []fs.DirEntry{}

	err := fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error walking through %s: %v\n", d.Name(), err)
			return err
		}
		if d.Name() == TerraformDir {
			return fs.SkipDir
		}

		if isTerraformProject(filesystem, d, path) {
			projects = append(projects, d)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
	return projects
}

func refreshProjects(m *model) {
	dir, err := os.Getwd()
	fmt.Println("Refreshing", dir)
	if err != nil {
		panic(err)
	}
	m.projects = findAllTerraformProjects(os.DirFS(dir))
	fmt.Printf("Found %d project(s):\n", len(m.projects))
	fmt.Println(m.projects)
}
