package main

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type terraformPlan struct {
	Add     int
	Change  int
	Destroy int
}

func updateAllPlans(projects []Project) tea.Cmd {
	return func() tea.Msg {
		for i := range projects {
			updatePlan(&projects[i])()
		}
		return refreshProjectsMsg(projects)
	}
}

func updatePlan(project *Project) tea.Cmd {
	return func() tea.Msg {
		planChannel := make(chan string)
		go runTerraformPlan(project.Path, planChannel)
		res := <-planChannel

		parsedPlan, err := parsePlanOutput(res)
		if err != nil {
			return errMsg{err}
		}
		project.TerraformPlan = parsedPlan

		return updatePlanMsg(*project)
	}
}

func runTerraformPlan(dir string, ch chan string) {
	defer close(ch)
	cmd := exec.Command("terraform", "plan")
	cmd.Dir = dir
	out, _ := cmd.CombinedOutput()
	ch <- string(out)
}

func parsePlanOutput(output string) (terraformPlan, error) {
	errors := strings.Contains(output, "Error:")
	if errors {
		return terraformPlan{-1, -1, -1}, nil
	}

	noChanges := strings.Contains(output, "No changes.")
	if noChanges {
		return terraformPlan{0, 0, 0}, nil
	}

	outputsOnly := strings.Count(output, "Changes to Outputs:")
	if outputsOnly > 0 {
		return terraformPlan{0, outputsOnly, 0}, nil
	}

	pattern := `(\d+) to add, (\d+) to change, (\d+) to destroy`
	regex := regexp.MustCompile(pattern)
	submatches := regex.FindStringSubmatch(output)

	toAdd, _ := strconv.Atoi(submatches[1])
	toChange, _ := strconv.Atoi(submatches[2])
	toDestroy, _ := strconv.Atoi(submatches[3])
	return terraformPlan{toAdd, toChange, toDestroy}, nil
}
