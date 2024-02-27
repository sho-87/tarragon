package main

import (
	"fmt"

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
	columnValid        = "Valid"
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
			HeaderStyle(tableHeader).
			SelectableRows(true).
			WithSelectedText("     ", "  •  ").
			Filtered(true).
			Focused(true).
			BorderRounded().
			SortByAsc(columnName).
			WithKeyMap(tableKeys).
			WithTargetWidth(winSize.Width).
			WithPageSize(winSize.Height - 5).
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

func generateRowsFromProjects(projects *[]Project) []table.Row {
	rows := []table.Row{}
	for i := range *projects {
		// FIXME: fix this mess
		addText := fmt.Sprint((*projects)[i].PlanChanges.Add)
		changeText := fmt.Sprint((*projects)[i].PlanChanges.Change)
		destroyText := fmt.Sprint((*projects)[i].PlanChanges.Destroy)
		if addText == PlanError.String() || changeText == PlanError.String() || destroyText == PlanError.String() {
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
			columnName:         (*projects)[i].Name,
			columnPath:         tablePath.Render((*projects)[i].Path),
			columnAdd:          addText,
			columnChange:       changeText,
			columnDestroy:      destroyText,
			columnValid:        validText,
			columnLastModified: (*projects)[i].LastModified.Format("2006-01-02 15:04:05"),
			columnProject:      (*projects)[i],
		})

		rows = append(rows, row)
	}

	return rows
}
