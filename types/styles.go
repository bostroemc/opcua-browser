package types

import "charm.land/lipgloss/v2"

type Styles struct {
	Index       lipgloss.Style
	Body        lipgloss.Style
	ActiveBody  lipgloss.Style
	Title       lipgloss.Style
	ActiveTitle lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Index:       lipgloss.NewStyle().Foreground(lipgloss.Color("#FAB387")).Bold(true),
		Body:        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("244")),
		ActiveBody:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("255")),
		Title:       lipgloss.NewStyle().Foreground(lipgloss.Color("244")).MarginTop(1).Italic(true),
		ActiveTitle: lipgloss.NewStyle().Foreground(lipgloss.Color("255")).MarginTop(1).Italic(true),
	}
}
