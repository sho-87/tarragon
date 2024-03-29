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
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/confirmation"
	tsize "github.com/kopoli/go-terminal-size"
)

var (
	versionFlag       bool
	version           string
	WinSize           tsize.Size
	SearchPath        string
	Debug             bool
	ValidateOnRefresh bool = true
)

type State int

const (
	tableView State = iota
	outputView
	confirmationView
)

type MainModel struct {
	err          error
	confirmation *confirmation.Model
	task         func(*MainModel) tea.Cmd
	help         help.Model
	message      string
	keys         KeyMap
	projects     []Project
	output       OutputModel
	spinner      spinner.Model
	table        TableModel
	progress     progress.Model
	percent      float64
	state        State
	working      bool
	refreshing   bool
}

type Project struct {
	LastModified time.Time
	Name         string
	Path         string
	LastAction   TerraformCommand
	Output       string
	Valid        string
	PlanChanges  TerraformChanges
}

type (
	UpdateValidateMsg  Project
	UpdatePlanMsg      Project
	UpdateApplyMsg     Project
	UpdatesFinishedMsg string
	RefreshFinishedMsg []Project
	ErrMsg             struct{ err error }
)

func (e ErrMsg) Error() string {
	return e.err.Error()
}

func initialModel() MainModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	table := createProjectsTable()
	output := OutputModel{width: WinSize.Width, height: WinSize.Height}
	output.createViewport()

	main := MainModel{
		state:        tableView,
		table:        table,
		output:       output,
		confirmation: createConfirmation(),
		keys:         mainKeys,
		help:         help.New(),
		spinner:      s,
		progress: progress.New(
			progress.WithGradient("#737c73", "#8992a7"),
			progress.WithWidth(WinSize.Width),
		),
		working:    false,
		refreshing: false,
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
				m.output.setTitle(project.Name, project.LastAction)
				m.output.viewport.SetContent(project.Output + strings.Repeat("\n", 4))
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

			if ValidateOnRefresh {
				m.refreshing = true
				m.working = true
				m.message = "Terraform Validate: all projects"

				var batchArgs []tea.Cmd
				batchArgs = append(batchArgs, m.spinner.Tick)
				for i := range len(m.projects) {
					batchArgs = append(batchArgs, runValidate(&m.projects[i]))
				}
				cmds = append(cmds, tea.Sequence(tea.Batch(batchArgs...), updatesFinished))
			}

		case UpdateValidateMsg:
			m.message = fmt.Sprintf("Validated %s", msg.Name)
			m.table.updateData(&m.projects)
			if m.refreshing {
				m.percent += float64(1) / float64(m.table.model.TotalRows())
			} else {
				m.percent += float64(1) / float64(len(m.table.model.SelectedRows()))
			}

		case UpdatePlanMsg:
			m.message = fmt.Sprintf("Updated %s", msg.Name)
			m.table.updateData(&m.projects)
			m.percent += float64(1) / float64(len(m.table.model.SelectedRows()))

		case UpdateApplyMsg:
			m.message = fmt.Sprintf("Applied %s", msg.Name)
			m.table.updateData(&m.projects)
			m.percent += float64(1) / float64(len(m.table.model.SelectedRows()))

		case UpdatesFinishedMsg:
			m.working = false
			m.refreshing = false
			m.message = string(msg)
			m.percent = 0.0

		case tea.KeyMsg:
			if !m.table.model.GetIsFilterInputFocused() {
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
					cmds = append(cmds, m.spinner.Tick, tea.Sequence(runValidate(highlightedProject), updatesFinished))

				case key.Matches(msg, m.keys.ValidateSelected):
					m.working = true
					m.message = "Terraform Validate: selected projects"

					var batchArgs []tea.Cmd
					batchArgs = append(batchArgs, m.spinner.Tick)
					for _, row := range m.table.model.SelectedRows() {
						project := matchProjectInMemory(row.Data[columnProject].(Project).Path, &m.projects)
						batchArgs = append(batchArgs, runValidate(project))
					}
					cmds = append(cmds, tea.Sequence(tea.Batch(batchArgs...), updatesFinished))

				case key.Matches(msg, m.keys.PlanHighlighted):
					m.working = true
					m.message = fmt.Sprintf("Terraform Plan: %s", project.Name)
					cmds = append(cmds, m.spinner.Tick, tea.Sequence(runPlan(highlightedProject), updatesFinished))

				case key.Matches(msg, m.keys.PlanSelected):
					m.working = true
					m.message = "Terraform Plan: selected projects"

					var batchArgs []tea.Cmd
					batchArgs = append(batchArgs, m.spinner.Tick)
					for _, row := range m.table.model.SelectedRows() {
						project := matchProjectInMemory(row.Data[columnProject].(Project).Path, &m.projects)
						batchArgs = append(batchArgs, runPlan(project))
					}
					cmds = append(cmds, tea.Sequence(tea.Batch(batchArgs...), updatesFinished))

				case key.Matches(msg, m.keys.ApplyHighlighted):
					m.task = func(m *MainModel) tea.Cmd {
						m.message = fmt.Sprintf("Terraform Apply: %s", project.Name)
						return tea.Sequence(runApply(highlightedProject), updatesFinished)
					}
					m.state = confirmationView

				case key.Matches(msg, m.keys.ApplySelected):
					m.task = func(m *MainModel) tea.Cmd {
						m.message = "Terraform Apply: selected projects"

						var batchArgs []tea.Cmd
						for _, row := range m.table.model.SelectedRows() {
							project := matchProjectInMemory(row.Data[columnProject].(Project).Path, &m.projects)
							batchArgs = append(batchArgs, runApply(project))
						}
						return tea.Sequence(tea.Batch(batchArgs...), updatesFinished)
					}
					m.state = confirmationView

				case key.Matches(msg, m.keys.SelectAll):
					rows := m.table.model.GetVisibleRows()
					for i, row := range rows {
						rows[i] = row.Selected(true)
					}

				case key.Matches(msg, m.keys.DeselectAll):
					m.table.model.WithAllRowsDeselected()
				}
			}
		}

	case confirmationView:
		msg, _ := msg.(tea.KeyMsg)
		switch {
		case key.Matches(msg, m.keys.Cancel):
			m.state = tableView
			m.working = false

		case key.Matches(msg, m.keys.No):
			m.state = tableView
			m.working = false

		case key.Matches(msg, m.keys.Yes):
			cmds = append(cmds, m.spinner.Tick, m.task(&m))
			m.state = tableView
			m.working = true

		default:
			_, cmd := m.confirmation.Update(msg)
			cmds = append(cmds, cmd)
		}

	case outputView:
		m.output.viewport, cmd = m.output.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m MainModel) renderProgress() string {
	working := ""
	progress := ""
	if m.working {
		working = fmt.Sprintf(" %s %s...", m.spinner.View(), m.message)
		if len(m.table.model.SelectedRows()) > 1 || m.refreshing {
			progress = m.progress.ViewAs(m.percent)
		}
	}
	return working + "\n" + progress
}

func (m MainModel) View() string {
	var output string

	switch m.state {
	case tableView:
		table := m.table.renderTable()
		progress := m.renderProgress()
		helpView := m.help.View(m.keys)

		contentHeight := lipgloss.Height(table) + lipgloss.Height(progress)
		paddingHeight := WinSize.Height - contentHeight - lipgloss.Height(helpView)

		output = table + progress + strings.Repeat("\n", max(paddingHeight, 0)) + helpView

	case confirmationView:
		table := m.table.renderTable()
		progress := m.renderProgress()
		confirm := m.confirmation.View()

		contentHeight := lipgloss.Height(table) + lipgloss.Height(progress)
		paddingHeight := WinSize.Height - contentHeight - lipgloss.Height(confirm)

		output = table + progress + strings.Repeat("\n", max(paddingHeight, 0)) + confirm

	case outputView:
		output = m.output.renderOutput()
	}
	return output
}

func main() {
	WinSize, _ = tsize.GetSize()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}

	flag.BoolVar(&versionFlag, "version", false, "Show version number")
	flag.BoolVar(&Debug, "debug", false, "Enable logging to file (debug.log)")
	flag.StringVar(&SearchPath, "path", cwd, "Path to search for Terraform projects")
	flag.Parse()

	if versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

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
	)
	_, runErr := p.Run()
	if runErr != nil {
		fmt.Printf("Uh oh, there was an error: %v\n", err)
		os.Exit(1)
	}
}
