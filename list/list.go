package list

import (
	"slices"
	"strings"
	"time"

	"github.com/bostroemc/tui/opcua-browser/types"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func New(id int, read chan types.OpcUaReadData, write chan types.DataPoint, data []types.DataPoint, update_rate int) Model {
	input := textinput.New()
	input.Placeholder = ""
	input.Prompt = "┃"
	input.CharLimit = 40

	return Model{Id: id, read: read, write: write, Data: data, UpdateRate: update_rate, Styles: types.DefaultStyles(), Input: input} //
}

type Model struct {
	Data   []types.DataPoint
	count  int //count increments every time Data is changed; count value is used to verify async operations
	Height int
	Width  int

	Autoupdate bool
	EditMode   bool
	UpdateRate int
	Status     string

	DataPoint types.DataPoint
	Input     textinput.Model

	read     chan types.OpcUaReadData
	write    chan types.DataPoint
	Styles   types.Styles
	index    int
	min, max int
	Id       int
	Active   int
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(readCmd(m.read), tickCmd(10*time.Second))
}

func (m Model) View() string {
	var s strings.Builder
	for i, l := range m.Data {
		node := nodeStyle.Render(l.Node)

		var value string

		if m.Active == m.Id && m.index == i {
			if m.EditMode {
				value = m.Styles.Edit.Render(m.Input.View())
			} else {
				value = valueStyle.Render(l.String())
			}

			gapWidth := m.Width - lipgloss.Width(node) - lipgloss.Width(value) - 0
			if gapWidth < 0 {
				gapWidth = 0
			}
			gap := strings.Repeat(" ", gapWidth)

			s.WriteString(m.Styles.Index.Render(lipgloss.JoinHorizontal(lipgloss.Top, node, gap, value)) + "\n")
		} else {
			value := valueStyle.Render(l.String())

			gapWidth := m.Width - lipgloss.Width(node) - lipgloss.Width(value) - 0
			if gapWidth < 0 {
				gapWidth = 0
			}
			gap := strings.Repeat(" ", gapWidth)

			s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, node, gap, value) + "\n")
		}
	}

	if m.Active == m.Id {
		m.Styles.ActiveBody = m.Styles.ActiveBody.Width(m.Width).Height(m.Height)
		return m.Styles.ActiveBody.Render(s.String())
	}

	m.Styles.Body = m.Styles.Body.Width(m.Width).Height(m.Height)
	return m.Styles.Body.Render(s.String())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Active != m.Id {
			m.EditMode = false
			break
		}
		if keyAction, ok := types.KeyActions[msg.String()]; ok {
			switch keyAction.Action {
			case "move_down":
				if !m.EditMode {
					m.index++
					m.index = min(m.index, len(m.Data)-1)
					m.SetMinMax(m.min, m.max)
				}
			case "move_up":
				if !m.EditMode {
					m.index--
					m.index = max(m.index, 0)
					m.SetMinMax(m.min, m.max)
				}
			case "delete":
				if !m.EditMode {
					m.Data = slices.Delete(m.Data, m.index, m.index+1)
					m.Increment()

					m.index = min(m.index, len(m.Data)-1)
					m.index = max(m.index, 0)
				}
			case "toggle_edit_mode":
				m.EditMode = !m.EditMode
				if m.EditMode {
					m.Input.SetValue(m.Data[m.index].String())
				}
			case "select":
				if m.EditMode {
					s := m.DataPoint

					s.SetPending(m.Input.Value())
					m.write <- s
					m.EditMode = false
					// m.Input.SetValue("")
				}
			}
		}

	case tickMsg:
		if m.Autoupdate {
			_data := make([]types.DataPoint, len(m.Data))
			copy(_data, m.Data) // Use the built-in copy function for slices

			m.read <- types.OpcUaReadData{Data: _data, Count: m.count}
		}
		return m, tickCmd(time.Duration(m.UpdateRate) * time.Millisecond)

	case readMsg:
		if m.count == msg.Count {
			for i, v := range msg.Data {
				if m.Data[i].Node == msg.Data[i].Node {
					m.Data[i].Value = v.Value
				}
			}
		}
		return m, readCmd(m.read)

	}

	m.Status = "autoupdate: OFF"
	if m.Autoupdate {
		m.Status = "autoupdate: ON"
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

func (m *Model) Increment() {
	m.count = (m.count + 1) % 10000
}

func (m *Model) SetView(height, width int) {
	m.Height, m.Width = height, 3*width/4
}

func (m *Model) SetMinMax(minimum, maximum int) {
	if m.index < minimum {
		m.min = m.index
		m.max = m.index + m.Height - 1
		m.max = min(m.max, len(m.Data)-1)
	}
	if m.index > maximum {
		m.max = m.index
		m.min = m.index - m.Height + 1
		m.min = max(m.min, 0)
	}
}
func (m Model) ActiveDataPoint() types.DataPoint {
	if len(m.Data) == 0 {
		return types.DataPoint{}
	}
	return m.Data[m.index]
}

type readMsg types.OpcUaReadData

func readCmd(read chan types.OpcUaReadData) tea.Cmd {
	return func() tea.Msg {
		return readMsg(<-read)
	}
}

type tickMsg time.Time

func tickCmd(t time.Duration) tea.Cmd {
	return tea.Every(t, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

var nodeStyle = lipgloss.NewStyle().
	Align(lipgloss.Left)

var valueStyle = lipgloss.NewStyle().
	Align(lipgloss.Right)

	// var editStyle = lipgloss.NewStyle().
	// 	Background(lipgloss.Color("#333333")). //Light blue
	// 	Foreground(lipgloss.Color("#FFFFFF")). // White
	// 	Padding(0, 1).
	// 	Align(lipgloss.Center)
