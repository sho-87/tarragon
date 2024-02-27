package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	Plan          TerraformCommand = "plan"
	Validate      TerraformCommand = "validate"
	Apply         TerraformCommand = "apply"
	PlanError     TerraformError   = -1
	DriftError    TerraformError   = -2
	ConfigValid   string           = "✓"
	ConfigInvalid string           = "✗"
	ConfigUnknown string           = "?"
)

type TerraformCommand string

func (c TerraformCommand) String() string {
	return string(c)
}

type TerraformError int

func (e TerraformError) Value() int {
	return int(e)
}

func (e TerraformError) String() string {
	return fmt.Sprint(e.Value())
}

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

type RegexMatchError struct {
	Message string
}

func (e RegexMatchError) Error() string {
	return e.Message
}

func updatesFinished() tea.Msg {
	return UpdatesFinishedMsg("Projects updated")
}

func updateValidate(project *Project) tea.Cmd {
	return func() tea.Msg {
		output := runTerraformCommand(project.Path, Validate)
		project.Output = output
		if strings.Contains(output, "The configuration is valid") {
			project.Valid = ConfigValid
		} else {
			project.Valid = ConfigInvalid
		}

		return UpdateValidateMsg(*project)
	}
}

func updatePlan(project *Project) tea.Cmd {
	return func() tea.Msg {
		output := runTerraformCommand(project.Path, Plan)
		parsedPlan := parsePlanOutput(output)
		project.TerraformPlan = parsedPlan
		project.Output = output

		return UpdatePlanMsg(*project)
	}
}

func runTerraformCommand(dir string, command TerraformCommand) string {
	cmd := exec.Command("terraform", command.String())
	cmd.Dir = dir

	out, _ := cmd.CombinedOutput()
	return string(out)
}

func parsePlanOutputJSON(output string) TerraformChanges {
	// Deprecated.
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
			return TerraformChanges{PlanError.Value(), PlanError.Value(), PlanError.Value()}
		} else if entry.Level == "info" && entry.Changes != (ChangeSummary{}) {
			changes := entry.Changes
			return TerraformChanges{changes.Add, changes.Change, changes.Remove}
		}
	}

	return TerraformChanges{PlanError.Value(), PlanError.Value(), PlanError.Value()}
}

func parsePlanOutput(output string) TerraformChanges {
	switch {
	case strings.Contains(output, "Error:"):
		return TerraformChanges{PlanError.Value(), PlanError.Value(), PlanError.Value()}
	case strings.Contains(output, "Objects have changed outside of Terraform"):
		return TerraformChanges{DriftError.Value(), DriftError.Value(), DriftError.Value()}
	case strings.Contains(output, "No changes."):
		return TerraformChanges{0, 0, 0}
	default:
		changes, err := regexMatchChanges(output)
		if err != nil {
			if Debug {
				log.Printf("Error parsing plan output: %s", err)
			}
			panic(err)
		}
		return changes
	}
}

func regexMatchChanges(output string) (TerraformChanges, error) {
	output = removeANSIEscapeCodes(output)
	re := regexp.MustCompile(`Plan: (\d+) to add, (\d+) to change, (\d+) to destroy.`)
	matches := re.FindStringSubmatch(output)

	if len(matches) == 4 {
		add, errAdd := strconv.Atoi(matches[1])
		change, errChange := strconv.Atoi(matches[2])
		destroy, errDestroy := strconv.Atoi(matches[3])

		if errAdd == nil && errChange == nil && errDestroy == nil {
			return TerraformChanges{Add: add, Change: change, Destroy: destroy}, nil
		}
	}
	return TerraformChanges{}, RegexMatchError{output}
}

func removeANSIEscapeCodes(input string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[mGKHF]`)
	return re.ReplaceAllString(input, "")
}
