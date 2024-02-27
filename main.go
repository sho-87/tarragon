package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var Debug bool = false
var SearchPath string

type State int

const (
	tableView State = iota
	outputView
)

type MainModel struct {
	state    State
	table    TableModel
	output   OutputModel
	projects []Project
	spinner  spinner.Model
	progress progress.Model
	working  bool
	err      error
	message  string
	keys     KeyMap
	help     help.Model
}

type Project struct {
	Name          string
	Path          string
	LastModified  time.Time
	Output        string
	TerraformPlan TerraformChanges
	Valid         string
}

type UpdateValidateMsg Project
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
	main := MainModel{
		state:    tableView,
		table:    table,
		keys:     mainKeys,
		help:     help.New(),
		spinner:  s,
		progress: progress.New(progress.WithDefaultGradient()),
		working:  false,
	}
	return main
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(tea.SetWindowTitle("tarragon"), m.spinner.Tick, refreshProjects)
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	project, _ := m.table.model.HighlightedRow().Data[columnProject].(Project)
	highlightedProject := matchProjectInMemory(project.Path, &m.projects)

	switch msg := msg.(type) {
	case ErrMsg:
		m.err = msg
		fmt.Printf("Error: %v\n", msg)
		cmds = append(cmds, tea.Quit)
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.ToggleOutput) {
			if m.state == outputView {
				m.state = tableView
			} else {
				m.output = OutputModel{title: highlightedProject.Name, content: highlightedProject.Output, width: 90, height: 25}
				m.state = outputView
			}
		}
	}

	switch m.state {
	case tableView:
		m.table.updateFooter()
		m.table.model, cmd = m.table.model.Update(msg)
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

		case UpdateValidateMsg:
			m.message = fmt.Sprintf("Validated %s", msg.Name)
			m.table.updateData(&m.projects)
			cmd := m.progress.IncrPercent(float64(1) / float64(len(m.table.model.SelectedRows())))
			cmds = append(cmds, cmd)

		case UpdatesFinishedMsg:
			m.working = false
			m.message = string(msg)

		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keys.Help):
				m.help.ShowAll = !m.help.ShowAll

			case key.Matches(msg, m.keys.Quit):
				cmds = append(cmds, tea.Quit)

			case key.Matches(msg, m.keys.Refresh):
				m.working = true
				cmds = append(cmds, m.spinner.Tick, refreshProjects)

			case key.Matches(msg, m.keys.ValidateHighlighted):
				m.working = true
				m.message = fmt.Sprintf("Terraform Validate: %s", project.Name)
				cmds = append(cmds, m.spinner.Tick, updateValidate(highlightedProject))

			case key.Matches(msg, m.keys.ValidateSelected):
				m.working = true
				m.message = "Terraform Validate: selected projects"

				var batchArgs []tea.Cmd
				batchArgs = append(batchArgs, m.spinner.Tick)
				for _, row := range m.table.model.SelectedRows() {
					project := matchProjectInMemory(row.Data[columnProject].(Project).Path, &m.projects)
					batchArgs = append(batchArgs, updateValidate(project))
				}
				cmds = append(cmds, tea.Sequence(tea.Batch(batchArgs...), updatesFinished))

			case key.Matches(msg, m.keys.PlanHighlighted):
				m.working = true
				m.message = fmt.Sprintf("Terraform Plan: %s", project.Name)
				cmds = append(cmds, m.spinner.Tick, updatePlan(highlightedProject))

			case key.Matches(msg, m.keys.PlanSelected):
				m.working = true
				m.message = "Terraform Plan: selected projects"

				var batchArgs []tea.Cmd
				batchArgs = append(batchArgs, m.spinner.Tick)
				for _, row := range m.table.model.SelectedRows() {
					project := matchProjectInMemory(row.Data[columnProject].(Project).Path, &m.projects)
					batchArgs = append(batchArgs, updatePlan(project))
				}
				cmds = append(cmds, tea.Sequence(tea.Batch(batchArgs...), updatesFinished))

			case key.Matches(msg, m.keys.SelectAll):
				rows := m.table.model.GetVisibleRows()
				for i, row := range rows {
					rows[i] = row.Selected(true)
				}

			case key.Matches(msg, m.keys.DeselectAll):
				m.table.model.WithAllRowsDeselected()
			}
		}

	case outputView:
		m.output, cmd = m.output.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) View() string {
	var output string

	switch m.state {
	case tableView:
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

		helpView := m.help.View(m.keys)

		output = body.String() + working + progress + strings.Repeat("\n", 10) + helpView
	case outputView:
		output = m.output.View()
	}
	return output
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

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, runErr := p.Run()
	if runErr != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
