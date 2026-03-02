package types

import "charm.land/lipgloss/v2"

type Styles struct {
	Index      lipgloss.Style
	Edit       lipgloss.Style
	Body       lipgloss.Style
	ActiveBody lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Index:      lipgloss.NewStyle().Foreground(lipgloss.Color("#FAB387")).Bold(true),
		Edit:       lipgloss.NewStyle().Foreground(lipgloss.Color("#FCAA95")).Bold(true),
		Body:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("244")),
		ActiveBody: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("255")),
	}
}
