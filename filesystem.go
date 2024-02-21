package main

import (
	"io/fs"
	"os"
	"path/filepath"

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

func findAllTerraformProjects(filesystem fs.FS) ([]Project, error) {
	projects := []Project{}

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
			info, err := d.Info()
			if err != nil {
				return err
			}
			project := Project{
				Name:         d.Name(),
				Path:         filepath.Join(SearchPath, path),
				LastModified: info.ModTime(),
			}
			projects = append(projects, project)
		}

		return nil
	})

	return projects, nil
}

func refreshProjects() tea.Msg {
	projects, err := findAllTerraformProjects(os.DirFS(SearchPath))
	if err != nil {
		return errMsg{err}
	}

	return refreshProjectsMsg(projects)
}
