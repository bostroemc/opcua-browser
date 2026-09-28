package overlaymethod

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	_ "charm.land/lipgloss/v2"
	"github.com/bostroemc/tui/opcua-browser/types"
	"github.com/gopcua/opcua/ua"
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

	Objects []*ua.ExtensionObject

	Data   []types.OpcUaInputArgumentData
	Styles types.Styles
	Show   bool

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
				m.index = min(m.index, len(m.Data)-1)
				m.SetMinMax(m.min, m.max)
			case "move_up":
				m.index--
				m.index = max(m.index, 0)
				m.SetMinMax(m.min, m.max)

			}
		}
	}

	return m, nil
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
		for _, o := range m.Data {
			arg, _ := o.ExtensionObject.Value.(*ua.Argument)
			s.WriteString(arg.Name + "\n")
			if arg.DataType.Namespace() != 0 {
				for _, field := range o.StructureDefinition.Fields {
					s.WriteString("  " + field.Name + "\n")
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
