package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type OutputModel struct {
	title    string
	action   TerraformCommand
	viewport viewport.Model
	width    int
	height   int
}

func (m *OutputModel) createViewport() {
	vpHeaderHeight := lipgloss.Height(m.outputHeader())
	vp := viewport.New(m.width, m.height-vpHeaderHeight*2)
	vp.YPosition = vpHeaderHeight + 1
	m.viewport = vp
}

func (m *OutputModel) setTitle(title string, lastAction TerraformCommand) {
	m.title = title
	m.action = lastAction
}

func (m *OutputModel) outputHeader() string {
	title := outputTitle.Render(fmt.Sprintf("Output (%s): %s", m.action, m.title))
	line := strings.Repeat("-", max(0, m.width-lipgloss.Width(title)))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, line)
	return header
}

func (m *OutputModel) outputFooter() string {
	info := outputInfo.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("-", max(0, m.width-lipgloss.Width(info)))
	footer := lipgloss.JoinHorizontal(lipgloss.Center, line, info)
	return footer
}

func (m *OutputModel) renderOutput() string {
	return fmt.Sprintf("%s\n%s\n%s", m.outputHeader(), m.viewport.View(), m.outputFooter())
}
