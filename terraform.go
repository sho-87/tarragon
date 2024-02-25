package main

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type TerraformCommand string

func (c TerraformCommand) String() string {
	return string(c)
}

const (
	Plan     TerraformCommand = "plan"
	Validate TerraformCommand = "validate"
	Apply    TerraformCommand = "apply"
)

type TerraformChanges struct {
	Add     int
	Change  int
	Destroy int
}

type PlanLogEntry struct {
	Level   string        `json:"@level"`
	Changes ChangeSummary `json:"changes"`
}

type ChangeSummary struct {
	Add    int `json:"add"`
	Change int `json:"change"`
	Remove int `json:"remove"`
}

func updatesFinished() tea.Msg {
	return UpdatesFinishedMsg("Projects updated")
}

func updatePlan(project *Project) tea.Cmd {
	if Debug {
		log.Printf("updatePlan: %p", project)
	}

	return func() tea.Msg {
		output := runTerraformCommand(project.Path, Plan)
		parsedPlan := parsePlanOutput(output)
		project.TerraformPlan = parsedPlan

		return UpdatePlanMsg(*project)
	}
}

func runTerraformCommand(dir string, command TerraformCommand) string {
	cmd := exec.Command("terraform", command.String(), "--json")
	cmd.Dir = dir

	// terraform plan errors also go to stdout and we want to capture those when parsing the plan output instead of here
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func parsePlanOutput(output string) TerraformChanges {
	logBuffer := []PlanLogEntry{}
	for _, line := range strings.Split(output, "\n") {
		var entry PlanLogEntry
		if err := json.NewDecoder(strings.NewReader(line)).Decode(&entry); err != nil {
			continue
		} else {
			logBuffer = append(logBuffer, entry)
		}
	}

	// iterate backwards because plan errors can come after changes
	// and we want to be alerted to errors instead in those cases
	for i := len(logBuffer) - 1; i >= 0; i-- {
		entry := logBuffer[i]

		if entry.Level == "error" {
			return TerraformChanges{-1, -1, -1}
		} else if entry.Level == "info" && entry.Changes != (ChangeSummary{}) {
			changes := entry.Changes
			return TerraformChanges{changes.Add, changes.Change, changes.Remove}
		}
	}

	return TerraformChanges{0, 0, 0}
}
