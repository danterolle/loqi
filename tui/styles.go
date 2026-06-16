package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	inputStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
	outputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
