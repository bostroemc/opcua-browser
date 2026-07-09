package address //address space

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bostroemc/tui/opcua-browser/types"
	"github.com/gopcua/opcua/ua"
)

func New(id int, browse chan types.OpcUaBrowserData, root string) Model {
	node, _ := ua.ParseNodeID(root)
	return Model{Id: id, browse: browse, Styles: types.DefaultStyles(), Node: node, stack: newStack(), Path: root} // "ns=8;s=plc/app/Application/sym"}
}

type Model struct {
	index  int
	browse chan types.OpcUaBrowserData
	Path   string
	// Active   int
	Node     *ua.NodeID
	Parent   types.Node
	Children []types.Node
	// id       int
	max    int // (min, max) chosen to obey these rules: 0<=min<=max<=len(Children); min<= index<= max; max-min <= Height
	min    int
	Height int
	Width  int
	stack  stack
	Styles types.Styles

	Id     int //module ID - module here refers to the various panes: navigation, value, status, ...
	Active int //active module ID
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(browseCmd(m.browse), initCmd(m.browse))
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.Active != m.Id {
			break
		}
		if keyAction, ok := types.KeyActions[msg.String()]; ok {
			switch keyAction.Action {
			case "move_down":
				m.index++
				m.index = min(m.index, len(m.Children)-1)
				m.SetMinMax(m.min, m.max)

			case "move_up":
				m.index--
				m.index = max(m.index, 0)
				m.SetMinMax(m.min, m.max)

			case "select":
				if len(m.Children) == 0 {
					break
				}
				m.Path = m.Children[m.index].NodeID.String()

				m.stack.Push(m.Node)
				m.Node = m.Children[m.index].NodeID

				m.browse <- types.OpcUaBrowserData{Node: m.Node}
				m.index = 0
				m.max = max(m.max, m.Height-5)
				m.SetMinMax(m.min, m.max)
			case "back":
				id := m.stack.Pop()
				if id == nil {
					break
				}
				m.index = 0
				m.Node = id
				m.Path = id.String()
				m.browse <- types.OpcUaBrowserData{Node: id}
				m.SetMinMax(m.min, m.max)
			case "root", "user_defined_node":
				m.stack.Clear()
				m.index = 0
				m.Path = *keyAction.Params.NodeId
				m.Node, _ = ua.ParseNodeID(m.Path)
				m.browse <- types.OpcUaBrowserData{Node: m.Node}

				m.SetMinMax(m.min, m.max)
			}
		}

	case browseMsg:
		m.Parent = msg.Parent
		m.Children = msg.Children
		return m, browseCmd(m.browse)
	}
	return m, nil
}

func (m Model) View() string {
	if len(m.Children) == 0 {
	}
	var s strings.Builder
	s.WriteString(m.Parent.BrowseName + "\n")
	for i, c := range m.Children {
		if i < m.min || i > m.max {
			continue
		}
		if m.Active == m.Id && m.index == i {
			s.WriteString("  " + m.Styles.Index.Render(c.BrowseName) + m.Styles.Index.Render(c.DataType) + "\n")
		} else {
			// s.WriteString("  " + c.BrowseName + c.DataType + " " + c.NodeClass.String() + "\n")
			s.WriteString("  " + c.BrowseName + c.DataType + "\n")
		}

	}

	if m.Active == m.Id {
		m.Styles.ActiveBody = m.Styles.ActiveBody.Width(m.Width).Height(m.Height)
		return lipgloss.JoinVertical(lipgloss.Left, m.Styles.ActiveTitle.Render(" address space"), m.Styles.ActiveBody.Render(s.String()))
	}

	m.Styles.Body = m.Styles.Body.Width(m.Width).Height(m.Height)
	return lipgloss.JoinVertical(lipgloss.Left, m.Styles.Title.Render(" address space"), m.Styles.Body.Render(s.String()))

}
func initCmd(browse chan types.OpcUaBrowserData) tea.Cmd {
	return func() tea.Msg {
		id, _ := ua.ParseNodeID("i=84") //"ns=8;s=plc/app/Application/sym")
		browse <- types.OpcUaBrowserData{Node: id}
		return nil
	}
}

func browseCmd(browse chan types.OpcUaBrowserData) tea.Cmd {
	return func() tea.Msg {
		return browseMsg(<-browse)
	}
}

type browseMsg types.OpcUaBrowserData

type stack struct {
	Push   func(*ua.NodeID)
	Pop    func() *ua.NodeID
	Length func() int
	Clear  func()
	Peek   func() string
}

func newStack() stack {
	buffer := make([]*ua.NodeID, 0)

	return stack{
		Push: func(n *ua.NodeID) {
			buffer = append(buffer, n)
		},
		Pop: func() *ua.NodeID {
			if len(buffer) == 0 {
				return nil
			}
			n := buffer[len(buffer)-1]
			buffer = buffer[:len(buffer)-1]
			return n
		},
		Length: func() int {
			return len(buffer)
		},
		Clear: func() {
			buffer = nil
		},
		Peek: func() string {
			if len(buffer) == 0 {
				return ""
			}
			n := buffer[len(buffer)-1]
			return n.String()
		},
	}
}

func (m *Model) pushView(selected, minimum, maximum int) {
	// m.stack.Push(selected)
}

func (m *Model) SetView(height, width int) {
	m.Height, m.Width = height, 1*width/4
	if m.max > m.Height-5 || m.max == m.min {
		m.max = m.min + m.Height - 5 //add that m.max must not be greater then len(Children)
		// m.max = min(m.max, len(m.Children))
	}
	// log.Println("SetView: ", m.min, m.max, m.Height)
}

func (m Model) ActiveNode() types.Node {
	return m.Children[m.index]
}

func (m *Model) SetMinMax(minimum, maximum int) {
	if m.index < minimum {
		m.min = m.index
		m.max = m.index + m.Height - 5
		m.max = min(m.max, len(m.Children)-1)
	}
	if m.index > maximum {
		m.max = m.index
		m.min = m.index - m.Height + 5
		m.min = max(m.min, 0)
	}
	// log.Println("SetMinMax: ", m.index, m.min, m.max, m.Height)
}
