package main

import (
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
)

func createConfirmation() *confirmation.Model {
	text := []string{"Are you sure?", warning.Render("This will apply with auto-approve"), "..."}
	prompt := confirmation.New(strings.Join(text, " "), confirmation.Undecided)
	prompt.Template = confirmation.TemplateYN
	prompt.ResultTemplate = confirmation.ResultTemplateYN
	prompt.KeyMap.SelectYes = append(prompt.KeyMap.SelectYes, "y")
	prompt.KeyMap.SelectNo = append(prompt.KeyMap.SelectNo, "n")
	model := confirmation.NewModel(prompt)
	model.Init()
	return model
}
