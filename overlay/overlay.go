package overlay

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	_ "charm.land/lipgloss/v2"
	"github.com/bostroemc/tui/opcua-browser/types"
)

func New() Model {
	return Model{
		// Endpoint: endpoint,
		Styles: types.DefaultStyles(),
	}
}

type Model struct {
	// Status string
	// Endpoint  string
	// DataPoint types.DataPoint
	Styles types.Styles
	Show   bool
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if keyAction, ok := types.KeyActions[msg.String()]; ok {
			switch keyAction.Action {
			case "escape":
				m.Show = false
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.Show || true {

		var s strings.Builder
		for _, key := range types.MyConfig.Keybinds {
			var keys strings.Builder
			for i, t := range key.Keys {
				keys.WriteString(t)
				if i < len(key.Keys)-1 {
					keys.WriteString(", ")
				}
			}
			gapWidth := 50 - lipgloss.Width(key.Action) - lipgloss.Width(keys.String()) - 4
			if gapWidth < 0 {
				gapWidth = 0
			}
			gap := strings.Repeat(" ", gapWidth)

			s.WriteString(key.Action + gap + keys.String() + "\n")
		}
		return m.Styles.Overlay.Render(s.String()) // lipgloss.JoinHorizontal(lipgloss.Top, s.String())
	}

	return "----"
}
