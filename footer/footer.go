package footer

import (
	"strings"

	"github.com/bostroemc/tui/opcua-browser/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func New(endpoint string) Model {
	// ta := textinput.New()
	// ta.Placeholder = "..."
	// ta.Prompt = "┃ "
	// ta.CharLimit = 40
	//
	return Model{
		// Input:    ta,
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
	// Input    textinput.Model

	// write chan types.DataPoint
	// message *string
	// viewport viewport.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.EditMode {
			break
		}
		if keyAction, ok := types.KeyActions[msg.String()]; ok {
			switch keyAction.Action {

			case "select":
			}
		}

	}

	return m, nil
}

func (m Model) View() string {
	path := footerStyle.Render(m.Path)
	icon := symbolStyle.Render(m.Status)

	endpoint := endpointStyle.Render(m.Endpoint)

	gapWidth := m.Width - lipgloss.Width(path) - lipgloss.Width(endpoint) - lipgloss.Width(icon)
	if gapWidth < 0 {
		gapWidth = 0
	}
	gap := strings.Repeat(" ", gapWidth)
	return lipgloss.JoinHorizontal(lipgloss.Top, path, gap, icon, endpoint)
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
