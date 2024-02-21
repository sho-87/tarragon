package main

import (
	"fmt"
	"io/fs"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const TerraformDir = ".terraform"

func isTerraformProject(filesystem fs.FS, d fs.DirEntry, path string) (bool, error) {
	if !d.IsDir() {
		return false, nil
	}

	dir, err := fs.ReadDir(filesystem, path)
	if err != nil {
		return false, err
	}

	for _, currentDir := range dir {
		if currentDir.IsDir() && currentDir.Name() == TerraformDir {
			return true, nil
		}
	}
	return false, nil
}

func findAllTerraformProjects(filesystem fs.FS) ([]fs.DirEntry, error) {
	projects := []fs.DirEntry{}

	fs.WalkDir(filesystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == TerraformDir {
			return fs.SkipDir
		}

		isProject, err := isTerraformProject(filesystem, d, path)

		if err != nil {
			return err
		} else if isProject {
			projects = append(projects, d)
		}

		return nil
	})

	return projects, nil
}

func refreshProjects() tea.Msg {
	dir, err := os.Getwd()
	if err != nil {
		return errMsg{err}
	}

	projects, err := findAllTerraformProjects(os.DirFS(dir))
	if err != nil {
		return errMsg{err}
	}

	fmt.Printf("Found %d project(s):\n", len(projects))
	fmt.Println(projects)

	return refreshProjectsMsg(projects)
}
