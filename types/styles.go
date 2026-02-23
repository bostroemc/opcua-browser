package types

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	DisabledCursor lipgloss.Style
	Cursor         lipgloss.Style
	Symlink        lipgloss.Style
	Directory      lipgloss.Style
	File           lipgloss.Style
	DisabledFile   lipgloss.Style
	Permission     lipgloss.Style
	Index          lipgloss.Style
	Edit           lipgloss.Style
	Disabledindex  lipgloss.Style
	FileSize       lipgloss.Style
	EmptyDirectory lipgloss.Style
	Body           lipgloss.Style
	ActiveBody     lipgloss.Style
}

func DefaultStyles() Styles {
	return DefaultStylesWithRenderer(lipgloss.DefaultRenderer())
}

// DefaultStylesWithRenderer defines the default styling for the file picker,
// with a given Lip Gloss renderer.
func DefaultStylesWithRenderer(r *lipgloss.Renderer) Styles {
	return Styles{
		DisabledCursor: r.NewStyle().Foreground(lipgloss.Color("247")),
		Cursor:         r.NewStyle().Foreground(lipgloss.Color("212")),
		Symlink:        r.NewStyle().Foreground(lipgloss.Color("36")),
		Directory:      r.NewStyle().Foreground(lipgloss.Color("99")),
		File:           r.NewStyle(),
		DisabledFile:   r.NewStyle().Foreground(lipgloss.Color("243")),
		Disabledindex:  r.NewStyle().Foreground(lipgloss.Color("247")),
		Permission:     r.NewStyle().Foreground(lipgloss.Color("244")),
		Index:          r.NewStyle().Foreground(lipgloss.Color("#FAB387")).Bold(true),
		Edit:           r.NewStyle().Foreground(lipgloss.Color("#FAC498")).Bold(true),
		Body:           r.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("244")),
		ActiveBody:     r.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("255")),
	}
}
