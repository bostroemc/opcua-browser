package main

import (
	"flag"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	browser "github.com/bostroemc/tui/opcua-browser/browser"
	"github.com/bostroemc/tui/opcua-browser/footer"
	list "github.com/bostroemc/tui/opcua-browser/list"
	"github.com/bostroemc/tui/opcua-browser/types"
)

type model struct {
	browser browser.Model
	list    list.Model
	footer  footer.Model

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
	case tea.KeyPressMsg:
		if keyAction, ok := types.KeyActions[msg.String()]; ok {
			switch keyAction.Action {
			case "quit":
				m.quitting = true
				return m, tea.Quit
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
		m.browser.SetView(m.height-3, m.width)
		m.list.SetView(m.height-3, m.width)
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

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	var s strings.Builder
	// s.WriteString(" rymden software\n")
	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, m.browser.View(), m.list.View()) + "\n")
	s.WriteString(m.footer.View())

	v := tea.NewView(s.String())
	v.AltScreen = true
	return v
}

func main() {
	username := flag.String("username", "", "User credentials for OPC UA server")
	password := flag.String("password", "", "User credentials for OPC UA server")
	flag.Parse()

	types.MyConfig.Init(username, password)

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

	_, _ = tea.NewProgram(m).Run()
}
