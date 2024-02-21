package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// https://www.zackproser.com/blog/bubbletea-state-machine
// https://github.com/charmbracelet/bubbletea/issues/27

var SearchPath string

type model struct {
	spinner  spinner.Model
	working  bool
	cursor   int
	projects []Project
	err      error
}

type Project struct {
	Name         string
	Path         string
	LastModified time.Time
}

type refreshProjectsMsg []Project
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return model{spinner: s, working: false}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, refreshProjects)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.working {
			m.spinner, cmd = m.spinner.Update(msg)
		}
		return m, cmd

	case refreshProjectsMsg:
		m.projects = msg
		m.working = false
		return m, nil

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.working = true
			return m, tea.Batch(m.spinner.Tick, refreshProjects)
		case "s":
			m.working = !m.working
			return m, m.spinner.Tick
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	var working string
	if m.working {
		working = fmt.Sprintf("\n\n   %s working...\n\n", m.spinner.View())
	} else {
		working = ""
	}

	projects := "Terraform Projects:\n"
	for _, p := range m.projects {
		projects += fmt.Sprintf("\n  - %s // %s (%s)\n", p.Name, p.Path, p.LastModified.Format("2006-01-02 15:04:05"))
	}

	return projects + working
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
	flag.StringVar(&SearchPath, "path", cwd, "Path to search for Terraform projects")
	flag.Parse()

	p := tea.NewProgram(initialModel())
	_, runErr := p.Run()
	if runErr != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
