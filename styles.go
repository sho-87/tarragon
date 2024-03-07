package main

import "github.com/charmbracelet/lipgloss"

var (
	tableBase          = lipgloss.NewStyle().BorderForeground(lipgloss.Color("#737c73")).Foreground(lipgloss.Color("#DCD7BA")).Align(lipgloss.Left)
	tableHighlighted   = lipgloss.NewStyle().Foreground(lipgloss.Color("#181616")).Background(lipgloss.Color("#7a8382")).Bold(true)
	tableHeader        = lipgloss.NewStyle().Foreground(lipgloss.Color("#8ba4b0")).Bold(true)
	tableHeaderPrimary = lipgloss.NewStyle().Foreground(lipgloss.Color("#8992a7"))
	tablePath          = lipgloss.NewStyle().Italic(true).Faint(true)
	tableDate          = lipgloss.NewStyle().Faint(true)
	success            = lipgloss.NewStyle().Foreground(lipgloss.Color("#87a987")).Align(lipgloss.Center)
	errorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#c4746e")).Align(lipgloss.Center)
	warning            = lipgloss.NewStyle().Foreground(lipgloss.Color("#b6927b"))
	outputTitle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#8ba4b0")).Bold(true).BorderStyle(lipgloss.RoundedBorder()).Padding(0, 2)
	outputInfo         = lipgloss.NewStyle().Foreground(lipgloss.Color("#8ba4b0")).Bold(true).BorderStyle(lipgloss.RoundedBorder()).Padding(0, 2).Faint(true)
)
