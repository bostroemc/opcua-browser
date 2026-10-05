package overlaymethod

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	_ "charm.land/lipgloss/v2"
	"github.com/bostroemc/tui/opcua-browser/backend"
	"github.com/bostroemc/tui/opcua-browser/types"
	"github.com/gopcua/opcua/ua"
)

func New(id int, backend *backend.ServiceOpcUa) Model {
	input := textinput.New()
	input.Placeholder = ""
	input.Prompt = "┃"
	input.CharLimit = 40

	s := input.Styles()
	s.Focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#FCAA95")).Bold(true)
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#FCAA95"))
	s.Cursor.Color = lipgloss.Color("#FCAA95")

	input.SetStyles(s)

	return Model{
		// Endpoint: endpoint,
		Id:      id,
		Styles:  types.DefaultStyles(),
		Input:   input,
		Width:   60,
		backend: backend,
	}
}

type Model struct {
	// Status string
	// Endpoint  string
	// DataPoint types.DataPoint

	// Objects []*ua.ExtensionObject

	Parent types.Node
	Method types.Node

	Data   []types.OpcUaInputArgumentData
	Fields []string
	Values []string
	Styles types.Styles
	Show   bool
	Input  textinput.Model

	Id     int
	Active int

	EditMode bool
	Height   int
	Width    int
	index    int
	min, max int

	backend *backend.ServiceOpcUa
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.Active != m.Id {
			m.EditMode = false
			break
		}
		if keyAction, ok := types.KeyActions[msg.String()]; ok {
			switch keyAction.Action {
			case "escape":
				m.Show = false
			case "move_down":
				if !m.EditMode {
					m.index++
					m.index = min(m.index, len(m.Fields)-1)
					m.SetMinMax(m.min, m.max)
				}
			case "move_up":
				if !m.EditMode {
					m.index--
					m.index = max(m.index, 0)
					m.SetMinMax(m.min, m.max)
				}
			case "toggle_edit_mode":
				m.EditMode = !m.EditMode
				if m.EditMode {
					m.Input.SetValue(m.Values[m.index])
				}
			case "select":
				if m.EditMode {
					m.Values[m.index] = m.Input.Value()

					m.EditMode = false
					m.Input.SetValue("")
				}
			case "call":
				m.backend.Call(m.Parent.NodeID, m.Method.NodeID, m.Data, m.Values)
			}
		}
	}

	if m.EditMode {
		m.Input.Focus()
	} else {
		m.Input.Blur()
	}
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	if m.Show || true {

		var s strings.Builder
		for i, o := range m.Fields {

			// arg, _ := o.ExtensionObject.Value.(*ua.Argument)
			node := nodeStyle.Render(o)
			value := valueStyle.Render(m.Values[i])

			if m.index == i && m.EditMode {
				value = m.Input.View()
			}

			gapWidth := m.Width - lipgloss.Width(node) - lipgloss.Width(value) - 2
			if gapWidth < 0 {
				gapWidth = 0
			}
			gap := strings.Repeat(" ", gapWidth)

			if m.index == i {
				s.WriteString(m.Styles.Index.Render(lipgloss.JoinHorizontal(lipgloss.Top, node, gap, value)) + "\n")

			} else {
				s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, node, gap, value) + "\n")
			}
		}
		if s.Len() == 0 {
			s.WriteString("no input arguments")
		}

		if m.Active == m.Id {
			m.Styles.ActiveBody = m.Styles.ActiveBody.Width(m.Width).Height(m.Height)
			return lipgloss.JoinVertical(lipgloss.Left, m.Styles.ActiveTitle.Render(" "+m.Method.BrowseName), m.Styles.ActiveBody.Render(s.String()))
		}

		m.Styles.Body = m.Styles.Body.Width(m.Width).Height(m.Height)
		return lipgloss.JoinVertical(lipgloss.Left, m.Styles.Title.Render(" "+m.Method.BrowseName), m.Styles.Body.Render(s.String()))
	}
	return "----"
}

func (m *Model) SetMinMax(minimum, maximum int) { //TODO: Improve recalculation of max/min when window resizes.
	if m.index < minimum {
		m.min = m.index
		m.max = m.index + m.Height - 4
		m.max = min(m.max, len(m.Fields)-1)
	}
	if m.index > maximum {
		m.max = m.index
		m.min = m.index - m.Height + 4
		m.min = max(m.min, 0)
	}
	// log.Println("SetMinMax: ", m.index, m.min, m.max, m.Height)
}

func (m *Model) LineCount() int {
	i := 0

	for _, o := range m.Data {

		arg, _ := o.ExtensionObject.Value.(*ua.Argument)
		i++
		if arg.DataType.Namespace() != 0 {
			i += len(o.StructureDefinition.Fields)
		}
	}

	return i
}

var nodeStyle = lipgloss.NewStyle().
	Align(lipgloss.Left)

var valueStyle = lipgloss.NewStyle().
	Align(lipgloss.Right)
