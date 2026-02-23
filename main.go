package main

import (
	"strings"
	"time"

	browser "github.com/bostroemc/tui/opcua-browser/browser"
	"github.com/bostroemc/tui/opcua-browser/footer"
	list "github.com/bostroemc/tui/opcua-browser/list"
	"github.com/bostroemc/tui/opcua-browser/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	browser browser.Model
	list    list.Model
	footer  footer.Model
	// activeNode   types.Node
	path     string
	quitting bool
	err      error

	state int

	height int
	width  int
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.browser.Init(), m.list.Init(), m.footer.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if keyAction, ok := types.KeyActions[msg.String()]; ok {
			switch keyAction.Action {
			case "quit":
				m.quitting = true
				return m, tea.Quit
			case "toggle_edit_mode":
				m.list.EditMode = !m.list.EditMode
			case "toggle_autoupdate":
				m.list.Autoupdate = !m.list.Autoupdate
			case "push":
				if m.state == 0 {
					m.list.Data = append(m.list.Data, types.DataPoint{Node: m.browser.ActiveNode().NodeID.String()})
					m.list.Increment()
				}

			case "toggle_focus":
				m.state = (m.state + 1) % 2
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.browser.SetView(m.height-4, m.width-4)
		m.list.SetView(m.height-4, m.width-4)
	}

	//Distribute state (i.e. active window) to the underlying modules
	m.browser.Active = m.state
	m.list.Active = m.state

	var cmd [3]tea.Cmd

	m.browser, cmd[0] = m.browser.Update(msg)
	m.list, cmd[1] = m.list.Update(msg)
	m.list.DataPoint = m.list.ActiveDataPoint()

	m.footer, cmd[2] = m.footer.Update(msg)
	m.footer.Status = m.list.Status
	m.footer.Path = m.browser.Path
	m.footer.Width = m.width

	return m, tea.Batch(cmd[0], cmd[1], cmd[2])
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString(" rymden software\n")
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, m.browser.View(), m.list.View()) + "\n")
	s.WriteString(m.footer.View())
	return s.String()
}

func main() {
	types.MyConfig.Init()

	ch_browse := make(chan types.OpcUaBrowserData)
	ch_read := make(chan types.OpcUaReadData)
	ch_write := make(chan types.DataPoint)

	go opcuaClient(types.MyConfig, ch_browse, ch_read, ch_write)

	m := model{
		browser: browser.New(0, ch_browse, "i=84"),
		list:    list.New(1, ch_read, ch_write, []types.DataPoint{}, types.MyConfig.UpdateRate),
		footer:  footer.New(types.MyConfig.Server.Endpoint),
	}

	time.Sleep(1000 * time.Millisecond) //wait for OPC UA service connection

	_, _ = tea.NewProgram(&m, tea.WithAltScreen()).Run()
}
