package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// https://www.zackproser.com/blog/bubbletea-state-machine
// https://github.com/charmbracelet/bubbletea/issues/27

type model struct {
	spinner  spinner.Model
	working  bool
	cursor   int
	projects []fs.DirEntry
	err      error
}

type refreshProjectsMsg []fs.DirEntry
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
		projects += fmt.Sprintf("\n  - %s\n", p.Name())
	}

	return projects + working
}

func main() {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
