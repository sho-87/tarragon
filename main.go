package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var debug bool = false
var SearchPath string

type model struct {
	table    projectsTable
	spinner  spinner.Model
	working  bool
	projects []Project
	err      error
	message  string
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
	table := createProjectsTable()
	return model{table: table, spinner: s, working: false}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, refreshProjects)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	m.table.updateFooter()
	m.table.model, cmd = m.table.model.Update(msg)
	cmds = append(cmds, cmd)

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

		for i := range m.projects {
			log.Printf("refreshFinishedMsg: %p", &m.projects[i])
		}

		m.table.updateData(&m.projects)
		for i := range m.projects {
			log.Printf("full circle: %p", &m.projects[i])
		}

		return m, nil

	case updatePlanMsg:
		m.message = fmt.Sprintf("Updated %s", msg.Name)
		m.table.updateData(&m.projects)
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
			cmds = append(cmds, tea.Quit)
		case "r":
			m.working = true
			cmds = append(cmds, m.spinner.Tick, refreshProjects)
		case "p":
			project := m.table.model.HighlightedRow().Data[columnProject].(Project)
			highlightedProject := matchHighlightedProject(project.Path, &m.projects)
			m.working = true
			m.message = fmt.Sprintf("Terraform Plan: %s", project.Name)
			cmds = append(cmds, m.spinner.Tick, updatePlan(highlightedProject))
		case "P":
			m.working = true
			m.message = "Terraform Plan: all projects"

			var batchArgs []tea.Cmd
			batchArgs = append(batchArgs, m.spinner.Tick)
			for i := range len(m.projects) {
				batchArgs = append(batchArgs, updatePlan(&m.projects[i]))
			}
			cmds = append(cmds, tea.Sequence(tea.Batch(batchArgs...), updatesFinished))
		case "s":
			m.working = !m.working
			cmds = append(cmds, m.spinner.Tick)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	selected := []string{}
	for _, row := range m.table.model.SelectedRows() {
		selected = append(selected, row.Data[columnName].(string))
	}

	body := strings.Builder{}
	body.WriteString(m.table.model.View())
	body.WriteString("\n")

	var working string
	if m.working {
		working = fmt.Sprintf("\n   %s %s...\n\n", m.spinner.View(), m.message)
	} else {
		working = ""
	}

	return body.String() + working
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}

	flag.BoolVar(&debug, "debug", false, "Enable logging to file (debug.log)")
	flag.StringVar(&SearchPath, "path", cwd, "Path to search for Terraform projects")
	flag.Parse()

	if debug {
		log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	p := tea.NewProgram(initialModel())
	_, runErr := p.Run()
	if runErr != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
