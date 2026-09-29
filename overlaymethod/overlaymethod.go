package overlaymethod

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	_ "charm.land/lipgloss/v2"
	"github.com/bostroemc/tui/opcua-browser/types"
	"github.com/gopcua/opcua/ua"
)

func New() Model {
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
		Styles: types.DefaultStyles(),
		Input:  input,
		Width:  60,
	}
}

type Model struct {
	// Status string
	// Endpoint  string
	// DataPoint types.DataPoint

	Objects []*ua.ExtensionObject

	Data   []types.OpcUaInputArgumentData
	Styles types.Styles
	Show   bool
	Input  textinput.Model

	EditMode bool
	Height   int
	Width    int
	index    int
	min, max int
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
			case "move_down":
				m.index++
				m.index = min(m.index, m.LineCount()-1)
				m.SetMinMax(m.min, m.max)
			case "move_up":
				m.index--
				m.index = max(m.index, 0)
				m.SetMinMax(m.min, m.max)
			case "toggle_edit_mode":
				m.EditMode = !m.EditMode
				if m.EditMode {
					// m.Input.SetValue(m.Data[m.index].String())
				}

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
		// for _, o := range m.Objects {
		// arg, _ := o.Value.(*ua.Argument)
		// s.WriteString(arg.Name + "\n")

		// if arg.DataType.Namespace() != 0 {
		// 	typeDef, _ := m.backend.GetStructureDefinition(arg.DataType)
		// 	binaryEncodingID, _ := m.backend.FindBinaryEncodingID(arg.DataType)
		// 	fmt.Println(binaryEncodingID)
		// 	for _, field := range typeDef.Fields {
		// 		fmt.Printf("Field Name: %s | DataType ID: %s %s\n", field.Name, field.DataType.String(), resolveDataType(field.DataType))
		//
		// 	}
		// }
		// }
		i := 0
		for _, o := range m.Data {

			arg, _ := o.ExtensionObject.Value.(*ua.Argument)
			node := nodeStyle.Render(arg.Name)
			var value string

			if m.index == i {
				if m.EditMode {
					value = m.Input.View()
				} else {
					value = valueStyle.Render(arg.Name)
				}

				gapWidth := m.Width - lipgloss.Width(node) - lipgloss.Width(value) - 2
				if gapWidth < 0 {
					gapWidth = 0
				}
				gap := strings.Repeat(" ", gapWidth)

				s.WriteString(m.Styles.Index.Render(lipgloss.JoinHorizontal(lipgloss.Top, node, gap, value)) + "\n")

			} else {
				s.WriteString(arg.Name + "\n")
			}
			i++
			if arg.DataType.Namespace() != 0 {
				for _, field := range o.StructureDefinition.Fields {
					if m.index == i {
						s.WriteString("  " + m.Styles.Index.Render(field.Name) + "\n")
					} else {
						s.WriteString("  " + field.Name + "\n")
					}
					i++
				}
			}
		}

		return m.Styles.Overlay.Render(s.String()) // lipgloss.JoinHorizontal(lipgloss.Top, s.String())
	}

	return "----"
}

func (m *Model) SetMinMax(minimum, maximum int) { //TODO: Improve recalculation of max/min when window resizes.
	if m.index < minimum {
		m.min = m.index
		m.max = m.index + m.Height - 4
		m.max = min(m.max, len(m.Data)-1)
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
