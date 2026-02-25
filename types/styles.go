package types

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Index      lipgloss.Style
	Edit       lipgloss.Style
	Body       lipgloss.Style
	ActiveBody lipgloss.Style
}

func DefaultStyles() Styles {
	return DefaultStylesWithRenderer(lipgloss.DefaultRenderer())
}

// DefaultStylesWithRenderer defines the default styling for the file picker,
// with a given Lip Gloss renderer.
func DefaultStylesWithRenderer(r *lipgloss.Renderer) Styles {
	return Styles{
		Index:      r.NewStyle().Foreground(lipgloss.Color("#FAB387")).Bold(true),
		Edit:       r.NewStyle().Foreground(lipgloss.Color("#FCAA95")).Bold(true),
		Body:       r.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("244")),
		ActiveBody: r.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("255")),
	}
}
