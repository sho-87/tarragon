package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	ToggleOutput        key.Binding
	Help                key.Binding
	Quit                key.Binding
	Up                  key.Binding
	Down                key.Binding
	Filter              key.Binding
	Refresh             key.Binding
	Select              key.Binding
	SelectAll           key.Binding
	DeselectAll         key.Binding
	PlanHighlighted     key.Binding
	PlanSelected        key.Binding
	ValidateHighlighted key.Binding
	ValidateSelected    key.Binding
}

var mainKeys = KeyMap{
	ToggleOutput: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle output"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "down"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Select: key.NewBinding(
		key.WithKeys("space", "enter"),
		key.WithHelp("space", "select"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "select all"),
	),
	DeselectAll: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deselect all"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh projects"),
	),
	PlanHighlighted: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan"),
	),
	PlanSelected: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "plan: selected"),
	),
	ValidateHighlighted: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "validate"),
	),
	ValidateSelected: key.NewBinding(
		key.WithKeys("V"),
		key.WithHelp("V", "validate: selected"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ValidateHighlighted, k.PlanHighlighted, k.Up, k.Down, k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ValidateHighlighted, k.PlanHighlighted},
		{k.ValidateSelected, k.PlanSelected},
		{k.Select, k.SelectAll, k.DeselectAll},
		{k.ToggleOutput, k.Refresh, k.Filter},
		{k.Help, k.Quit},
	}
}
