package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

type TableModel struct {
	model table.Model
}

func matchProjectInMemory(path string, projects *[]Project) *Project {
	for i := range *projects {
		if (*projects)[i].Path == path {
			return &(*projects)[i]
		}
	}
	return nil
}

func (m *TableModel) updateData(projects *[]Project) {
	// FIXME: selected rows are lost when updating the table because all rows are replaced
	// FIXME: clearing a filter currently doesnt update the table to show all rows
	// https://github.com/Evertras/bubble-table/issues/136
	m.model = m.model.WithRows(generateRowsFromProjects(projects))
	m.updateFooter()
}

func (m *TableModel) updateFooter() {
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
	columnProject      = "Project"
)

func createProjectsTable() TableModel {
	columns := generateColumns()
	rows := generateRowsFromProjects(&[]Project{})

	tableKeys := table.DefaultKeyMap()
	tableKeys.RowDown.SetKeys(mainKeys.Down.Keys()...)
	tableKeys.RowUp.SetKeys(mainKeys.Up.Keys()...)
	tableKeys.Filter.SetKeys(mainKeys.Filter.Keys()...)

	model := TableModel{
		model: table.New(columns).
			WithRows(rows).
			HeaderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)).
			SelectableRows(true).
			WithSelectedText("   ", " • ").
			Filtered(true).
			Focused(true).
			BorderRounded().
			SortByAsc(columnLastModified).
			WithKeyMap(tableKeys).
			WithTargetWidth(100).
			WithMaxTotalWidth(200).
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

func generateRowsFromProjects(projects *[]Project) []table.Row {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	pathStyle := lipgloss.NewStyle().Italic(true).Faint(true)

	rows := []table.Row{}
	for i := range *projects {
		// FIXME: fix this mess
		addText := fmt.Sprint((*projects)[i].TerraformPlan.Add)
		changeText := fmt.Sprint((*projects)[i].TerraformPlan.Change)
		destroyText := fmt.Sprint((*projects)[i].TerraformPlan.Destroy)
		if addText == PlanError.String() || changeText == PlanError.String() || destroyText == PlanError.String() {
			addText = errorStyle.Render("Error")
			changeText = errorStyle.Render("Error")
			destroyText = errorStyle.Render("Error")
		} else if addText == DriftError.String() || changeText == DriftError.String() || destroyText == DriftError.String() {
			addText = errorStyle.Render("Drift")
			changeText = errorStyle.Render("Drift")
			destroyText = errorStyle.Render("Drift")
		}

		row := table.NewRow(table.RowData{
			columnName:         (*projects)[i].Name,
			columnPath:         pathStyle.Render((*projects)[i].Path),
			columnAdd:          addText,
			columnChange:       changeText,
			columnDestroy:      destroyText,
			columnLastModified: (*projects)[i].LastModified.Format("2006-01-02 15:04:05"),
			columnProject:      (*projects)[i],
		})

		rows = append(rows, row)
	}

	return rows
}
