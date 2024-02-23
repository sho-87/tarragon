package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
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
	message  string
}

type Project struct {
	Name          string
	Path          string
	LastModified  time.Time
	TerraformPlan terraformPlan
}

type updatePlanMsg Project
type updatesFinishedMsg string
type refreshFinishedMsg []Project
type errMsg struct{ err error }

func (e errMsg) Error() string {
	return e.err.Error()
}

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

	case refreshFinishedMsg:
		m.projects = msg
		m.working = false
		return m, nil

	case updatePlanMsg:
		m.message = fmt.Sprintf("Updated %s", msg.Name)
		return m, nil

	case updatesFinishedMsg:
		m.working = false
		m.message = string(msg)
		return m, nil

	case errMsg:
		m.err = msg
		fmt.Printf("Error: %v\n", msg)
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.working = true
			return m, tea.Batch(m.spinner.Tick, refreshProjects)
		case "u":
			m.working = true
			m.message = "Updating project"
			return m, tea.Batch(m.spinner.Tick, updatePlan(&m.projects[1]))
		case "U":
			m.working = true
			m.message = "Updating all projects"

			var batchArgs []tea.Cmd
			batchArgs = append(batchArgs, m.spinner.Tick)

			for i := 0; i < len(m.projects); i++ {
				batchArgs = append(batchArgs, updatePlan(&m.projects[i]))
			}
			return m, tea.Sequence(tea.Batch(batchArgs...), updatesFinished)
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
		working = fmt.Sprintf("\n\n   %s %s...\n\n", m.spinner.View(), m.message)
	} else {
		working = ""
	}

	projects := "Terraform Projects:\n"
	for _, p := range m.projects {
		projects += fmt.Sprintf("\n  - %s | %s | Add: %d Change: %d Destroy: %d | (%s)\n", p.Name, p.Path, p.TerraformPlan.Add, p.TerraformPlan.Change, p.TerraformPlan.Destroy, p.LastModified.Format("2006-01-02 15:04:05"))
	}

	goroutines := runtime.NumGoroutine()

	return projects + working + fmt.Sprintf("%d", goroutines)
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
