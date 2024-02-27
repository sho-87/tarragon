package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	Cancel              key.Binding
	ToggleOutput        key.Binding
	Help                key.Binding
	Quit                key.Binding
	Up                  key.Binding
	Down                key.Binding
	Yes                 key.Binding
	No                  key.Binding
	Filter              key.Binding
	Refresh             key.Binding
	Select              key.Binding
	SelectAll           key.Binding
	DeselectAll         key.Binding
	PlanHighlighted     key.Binding
	PlanSelected        key.Binding
	ValidateHighlighted key.Binding
	ValidateSelected    key.Binding
	ApplyHighlighted    key.Binding
	ApplySelected       key.Binding
}

var mainKeys = KeyMap{
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
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
	Yes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
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
	ApplyHighlighted: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "apply"),
	),
	ApplySelected: key.NewBinding(
		key.WithKeys("A"),
		key.WithHelp("A", "apply: selected"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ValidateHighlighted, k.PlanHighlighted, k.ApplyHighlighted, k.Up, k.Down, k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ValidateHighlighted, k.PlanHighlighted, k.ApplyHighlighted},
		{k.ValidateSelected, k.PlanSelected, k.ApplySelected},
		{k.Select, k.SelectAll, k.DeselectAll},
		{k.ToggleOutput, k.Refresh, k.Filter},
		{k.Help, k.Quit},
	}
}
