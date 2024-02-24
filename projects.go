package main

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

type Project struct {
	Name          string
	Path          string
	LastModified  time.Time
	TerraformPlan terraformPlan
}

type projectsTable struct {
	model table.Model
}

func (m *projectsTable) updateData(projects []Project) {
	// FIXME: selected rows are lost when updating the table because all rows are replaced
	// FIXME: clearing a filter currently doesnt update the table to show all rows
	// https://github.com/Evertras/bubble-table/issues/136
	m.model = m.model.WithRows(generateRowsFromProjects(projects))
	m.updateFooter()
}

func (m *projectsTable) updateFooter() {
	footerText := fmt.Sprintf(
		"Page %d/%d | # Projects: %d",
		m.model.CurrentPage(),
		m.model.MaxPages(),
		m.model.TotalRows(),
	)

	if m.model.GetIsFilterInputFocused() {
		footerText += fmt.Sprintf(" | Filter: %s", m.model.GetCurrentFilter())
	}

	m.model = m.model.WithStaticFooter(footerText)
}

const (
	columnName         = "ProjectName"
	columnPath         = "Path"
	columnAdd          = "Add"
	columnChange       = "Change"
	columnDestroy      = "Destroy"
	columnLastModified = "LastModified"
)

func createProjectsTable() projectsTable {
	columns := generateColumns()
	rows := generateRowsFromProjects([]Project{})

	keys := table.DefaultKeyMap()
	keys.RowDown.SetKeys("j", "down", "s")
	keys.RowUp.SetKeys("k", "up", "w")
	keys.Filter.SetKeys("/", "f")

	model := projectsTable{
		model: table.New(columns).
			WithRows(rows).
			HeaderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)).
			SelectableRows(true).
			WithSelectedText("   ", " * ").
			Filtered(true).
			Focused(true).
			BorderRounded().
			SortByAsc(columnLastModified).
			WithKeyMap(keys).
			WithTargetWidth(100).
			WithMaxTotalWidth(200).
			WithPageSize(10).
			WithMultiline(true).
			WithBaseStyle(
				lipgloss.NewStyle().
					BorderForeground(lipgloss.Color("#a38")).
					Foreground(lipgloss.Color("#a7a")).
					Align(lipgloss.Left),
			).
			HighlightStyle(
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("#88ff55")).
					Background(lipgloss.Color("#555055")),
			),
	}

	model.updateFooter()
	return model
}

func generateColumns() []table.Column {
	columns := []table.Column{
		table.NewFlexColumn(columnName, "Name", 1).
			WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#88f"))).
			WithFiltered(true),
		table.NewFlexColumn(columnPath, "Path", 3).WithFiltered(true),
		table.NewColumn(columnAdd, "Add", 10),
		table.NewColumn(columnChange, "Change", 10),
		table.NewColumn(columnDestroy, "Destroy", 10),
		table.NewColumn(columnLastModified, "Last Modified", 20),
	}

	return columns
}

func generateRowsFromProjects(projects []Project) []table.Row {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	rows := []table.Row{}
	for _, entry := range projects {
		addText := fmt.Sprint(entry.TerraformPlan.Add)
		if entry.TerraformPlan.Add == -1 {
			addText = errorStyle.Render("Error")
		}
		changeText := fmt.Sprint(entry.TerraformPlan.Change)
		if entry.TerraformPlan.Change == -1 {
			changeText = errorStyle.Render("Error")
		}
		destroyText := fmt.Sprint(entry.TerraformPlan.Destroy)
		if entry.TerraformPlan.Destroy == -1 {
			destroyText = errorStyle.Render("Error")
		}

		row := table.NewRow(table.RowData{
			columnName:         entry.Name,
			columnPath:         lipgloss.NewStyle().Italic(true).Faint(true).Render(entry.Path),
			columnAdd:          addText,
			columnChange:       changeText,
			columnDestroy:      destroyText,
			columnLastModified: entry.LastModified.Format("2006-01-02 15:04:05"),
		})
		rows = append(rows, row)
	}

	return rows
}
