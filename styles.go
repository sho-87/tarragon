package main

import "github.com/charmbracelet/lipgloss"

var (
	tableBase          = lipgloss.NewStyle().BorderForeground(lipgloss.Color("#a38")).Foreground(lipgloss.Color("#a7a")).Align(lipgloss.Left)
	tableHighlighted   = lipgloss.NewStyle().Foreground(lipgloss.Color("#88ff55")).Background(lipgloss.Color("#555055"))
	tableHeader        = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Bold(true)
	tableHeaderPrimary = lipgloss.NewStyle().Foreground(lipgloss.Color("#88f"))
	tablePath          = lipgloss.NewStyle().Italic(true).Faint(true)
	success            = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Align(lipgloss.Center)
	errorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Align(lipgloss.Center)
	warning            = lipgloss.NewStyle().Foreground(lipgloss.Color("#d97d0d"))
	outputTitle        = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 2)
	outputInfo         = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 2).Faint(true)
)
