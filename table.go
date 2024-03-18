package main

import (
	"fmt"
	"slices"
	"strings"

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
	var selected []string

	for _, row := range m.model.SelectedRows() {
		selected = append(selected, row.Data[columnProject].(Project).Path)
	}

	m.model = m.model.WithRows(generateRowsFromProjects(projects, selected))

	m.updateFooter()
}

func (m *TableModel) updateFooter() {
	footerText := fmt.Sprintf(
		"Page %d/%d  |  # Projects: %d",
		m.model.CurrentPage(),
		m.model.MaxPages(),
		m.model.TotalRows(),
	)

	m.model = m.model.WithStaticFooter(footerText)
}

const (
	columnName         = "ProjectName"
	columnPath         = "Path"
	columnAdd          = "Add"
	columnChange       = "Change"
	columnDestroy      = "Destroy"
	columnLastModified = "LastModified"
	columnValid        = "Valid"
	columnProject      = "Project"
)

func (m *TableModel) renderTable() string {
	filter := ""
	if m.model.GetIsFilterInputFocused() {
		filter = fmt.Sprintf(" Filter: %s_", m.model.GetCurrentFilter())
		filter = tableFilterTyping.Render(filter)
	} else if m.model.GetCurrentFilter() != "" {
		filter = fmt.Sprintf(" Filter: %s", m.model.GetCurrentFilter())
		filter = tableFilterSet.Render(filter)
	}

	body := strings.Builder{}
	body.WriteString("\n\n")
	body.WriteString(filter)
	body.WriteString("\n")
	body.WriteString(m.model.View())
	body.WriteString("\n\n")

	return body.String()
}

func createProjectsTable() TableModel {
	columns := generateColumns()
	rows := generateRowsFromProjects(&[]Project{}, []string{})

	tableKeys := table.DefaultKeyMap()
	tableKeys.RowDown.SetKeys(mainKeys.Down.Keys()...)
	tableKeys.RowUp.SetKeys(mainKeys.Up.Keys()...)
	tableKeys.Filter.SetKeys(mainKeys.Filter.Keys()...)

	model := TableModel{
		model: table.New(columns).
			WithRows(rows).
			HeaderStyle(tableHeader).
			SelectableRows(true).
			WithSelectedText("     ", "  •  ").
			Filtered(true).
			Focused(true).
			BorderRounded().
			SortByAsc(columnName).
			WithKeyMap(tableKeys).
			WithTargetWidth(WinSize.Width).
			WithPageSize(WinSize.Height - 5).
			WithMultiline(false).
			WithBaseStyle(tableBase).
			HighlightStyle(tableHighlighted),
	}

	model.updateFooter()
	return model
}

func generateColumns() []table.Column {
	columns := []table.Column{
		table.NewFlexColumn(columnName, "Name", 2).WithStyle(tableHeaderPrimary).WithFiltered(true),
		table.NewFlexColumn(columnPath, "Path", 4).WithFiltered(true),
		table.NewFlexColumn(columnValid, "Valid", 1),
		table.NewFlexColumn(columnAdd, "Add", 1),
		table.NewFlexColumn(columnChange, "Change", 1),
		table.NewFlexColumn(columnDestroy, "Destroy", 1),
		table.NewFlexColumn(columnLastModified, "Last Modified", 3),
	}

	return columns
}

func generateRowsFromProjects(projects *[]Project, selected []string) []table.Row {
	rows := []table.Row{}
	for i := range *projects {
		// FIXME: fix this mess
		addText := fmt.Sprint((*projects)[i].PlanChanges.Add)
		changeText := fmt.Sprint((*projects)[i].PlanChanges.Change)
		destroyText := fmt.Sprint((*projects)[i].PlanChanges.Destroy)
		if addText == PlanError.String() || changeText == PlanError.String() ||
			destroyText == PlanError.String() {
			addText = errorStyle.Render("Error")
			changeText = errorStyle.Render("Error")
			destroyText = errorStyle.Render("Error")
		} else if addText == DriftError.String() || changeText == DriftError.String() || destroyText == DriftError.String() {
			addText = errorStyle.Render("Drift")
			changeText = errorStyle.Render("Drift")
			destroyText = errorStyle.Render("Drift")
		}

		var validText string
		if (*projects)[i].Valid == ConfigValid {
			validText = success.Render(ConfigValid)
		} else if (*projects)[i].Valid == ConfigInvalid {
			validText = errorStyle.Render(ConfigInvalid)
		} else {
			validText = ConfigUnknown
		}

		row := table.NewRow(table.RowData{
			columnName:    (*projects)[i].Name,
			columnPath:    tablePath.Render((*projects)[i].Path),
			columnAdd:     addText,
			columnChange:  changeText,
			columnDestroy: destroyText,
			columnValid:   validText,
			columnLastModified: tableDate.Render(
				(*projects)[i].LastModified.Format("2006-01-02 15:04:05"),
			),
			columnProject: (*projects)[i],
		})

		if slices.Contains(selected, (*projects)[i].Path) {
			row = row.Selected(true)
		}

		rows = append(rows, row)
	}

	return rows
}
