package main

import (
	"fmt"
	"io/fs"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cursor   int
	projects []fs.DirEntry
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	refreshProjects(&m)
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		k := msg.(tea.KeyMsg)
		if k.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}

		keyString := k.String()
		switch keyString {
		case "q":
			return m, tea.Quit
		case "r":
			refreshProjects(&m)
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Hello, Bubble Tea!")
}

func main() {
	model := initialModel()
	p := tea.NewProgram(model)
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
