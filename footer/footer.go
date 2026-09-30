package footer

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bostroemc/tui/opcua-browser/types"
)

func New(endpoint string) Model {
	return Model{
		Endpoint: endpoint,
	}

}

type Model struct {
	Width     int
	Path      string
	Status    string
	Endpoint  string
	DataPoint types.DataPoint

	EditMode bool
	Action   string
	Message  string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.EditMode {
			break
		}
		if keyAction, ok := types.KeyActions[msg.String()]; ok {
			m.Action = keyAction.Action
			m.Message = msg.String()
		}

	}

	return m, nil
}

func (m Model) View() string {
	path := footerStyle.Render(m.Path)
	icon := symbolStyle.Render(m.Status)

	action := debugStyle.Render(m.Action)
	message := debugStyle.Render(m.Message)

	endpoint := endpointStyle.Render(m.Endpoint)

	gapWidth := m.Width - lipgloss.Width(path) - lipgloss.Width(endpoint) - lipgloss.Width(icon) - lipgloss.Width(action) - lipgloss.Width(message)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := strings.Repeat(" ", gapWidth)
	return lipgloss.JoinHorizontal(lipgloss.Top, path, action, message, gap, icon, endpoint)
}

var footerStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#89B4FA")). //TODO: move to package types
	Foreground(lipgloss.Color("#181825")).
	Padding(0, 1).
	BorderTop(false)

var endpointStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#89B4FA")).
	Foreground(lipgloss.Color("#181825")).
	Padding(0, 1).
	Align(lipgloss.Right)
var symbolStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#333333")).
	Foreground(lipgloss.Color("#FFFFFF")).
	Padding(0, 1).
	Align(lipgloss.Center)
var debugStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#FACF89")).
	Foreground(lipgloss.Color("#181825")).
	Padding(0, 1).
	Align(lipgloss.Center)
