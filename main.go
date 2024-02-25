package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var Debug bool = false
var SearchPath string

type MainModel struct {
	table     TableModel
	altscreen bool
	spinner   spinner.Model
	progress  progress.Model
	working   bool
	projects  []Project
	err       error
	message   string
}

type UpdatePlanMsg Project
type UpdatesFinishedMsg string
type RefreshFinishedMsg []Project
type ErrMsg struct{ err error }

func (e ErrMsg) Error() string {
	return e.err.Error()
}

func initialModel() MainModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	table := createProjectsTable()
	return MainModel{table: table, spinner: s, progress: progress.New(progress.WithDefaultGradient()), working: false}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, refreshProjects)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	m.table.updateFooter()
	m.table.model, cmd = m.table.model.Update(msg)
	project, _ := m.table.model.HighlightedRow().Data[columnProject].(Project)
	highlightedProject := matchProjectInMemory(project.Path, &m.projects)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.working {
			m.spinner, cmd = m.spinner.Update(msg)
		}
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case RefreshFinishedMsg:
		m.projects = msg
		m.working = false
		m.table.updateData(&m.projects)

	case UpdatePlanMsg:
		m.message = fmt.Sprintf("Updated %s", msg.Name)
		m.table.updateData(&m.projects)
		cmd := m.progress.IncrPercent(float64(1) / float64(len(m.table.model.SelectedRows())))
		cmds = append(cmds, cmd)

	case UpdatesFinishedMsg:
		m.working = false
		m.message = string(msg)

	case ErrMsg:
		m.err = msg
		fmt.Printf("Error: %v\n", msg)
		cmds = append(cmds, tea.Quit)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			cmds = append(cmds, tea.Quit)
		case "r":
			m.working = true
			cmds = append(cmds, m.spinner.Tick, refreshProjects)
		case "p":
			m.working = true
			m.message = fmt.Sprintf("Terraform Plan: %s", project.Name)
			cmds = append(cmds, m.spinner.Tick, updatePlan(highlightedProject))
		case "P":
			m.working = true
			m.message = "Terraform Plan: selected projects"

			var batchArgs []tea.Cmd
			batchArgs = append(batchArgs, m.spinner.Tick)
			for _, row := range m.table.model.SelectedRows() {
				project := matchProjectInMemory(row.Data[columnProject].(Project).Path, &m.projects)
				batchArgs = append(batchArgs, updatePlan(project))
			}
			cmds = append(cmds, tea.Sequence(tea.Batch(batchArgs...), updatesFinished))
		case "s":
			rows := m.table.model.GetVisibleRows()
			for i, row := range rows {
				rows[i] = row.Selected(true)
			}
		case "d":
			m.table.model.WithAllRowsDeselected()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
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

	var progress string
	if m.progress.Percent() > 0 && m.progress.Percent() < 1 {
		progress = fmt.Sprintf("\n%s\n", m.progress.View())
	} else {
		progress = ""
	}

	return body.String() + working + progress
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}

	flag.BoolVar(&Debug, "debug", false, "Enable logging to file (debug.log)")
	flag.StringVar(&SearchPath, "path", cwd, "Path to search for Terraform projects")
	flag.Parse()

	if Debug {
		log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, runErr := p.Run()
	if runErr != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
