package main

import (
	"flag"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	address "github.com/bostroemc/tui/opcua-browser/address"
	data "github.com/bostroemc/tui/opcua-browser/data"
	"github.com/bostroemc/tui/opcua-browser/footer"
	"github.com/bostroemc/tui/opcua-browser/overlay"
	"github.com/bostroemc/tui/opcua-browser/types"
)

type model struct {
	address address.Model
	data    data.Model
	footer  footer.Model
	overlay overlay.Model

	path     string
	quitting bool
	err      error
	info     bool

	state int

	height int
	width  int
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.address.Init(), m.data.Init(), m.footer.Init(), m.overlay.Init())
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
				m.data.Autoupdate = !m.data.Autoupdate
			case "push":
				if m.state == 0 {
					m.data.Data = append(m.data.Data, types.DataPoint{Node: m.address.ActiveNode().NodeID.String()})
					m.data.Increment()
					m.data.SetMinMax(0, len(m.data.Data)-1)
				}

			case "toggle_focus":
				m.state = (m.state + 1) % 2
			case "show_info":
				m.info = true
			case "hide_info":
				m.info = false
			}
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.address.SetView(m.height-3, m.width)
		m.data.SetView(m.height-3, m.width)
		// log.Println(m.height, m.width)  TODO: Delete OK
	}

	//Distribute state (i.e. active window) to the underlying modules
	m.address.Active = m.state
	m.data.Active = m.state

	var cmd [3]tea.Cmd

	m.address, cmd[0] = m.address.Update(msg)
	m.data, cmd[1] = m.data.Update(msg)
	m.data.DataPoint = m.data.ActiveDataPoint()

	m.footer, cmd[2] = m.footer.Update(msg)
	m.footer.Status = m.data.Status
	m.footer.Path = m.address.Path
	m.footer.Width = m.width

	return m, tea.Batch(cmd[0], cmd[1], cmd[2])
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	var s strings.Builder

	s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, m.address.View(), m.data.View()) + "\n")
	s.WriteString(m.footer.View())

	if m.info {
		baseLayer := lipgloss.NewLayer(s.String()).ID("base")
		overlay := lipgloss.NewLayer(m.overlay.View()).ID("overlay").X(40).Y(10).Z(10)
		compositor := lipgloss.NewCompositor(baseLayer, overlay)
		v := tea.NewView(compositor.Render())
		v.AltScreen = true
		return v
	}

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

	// f, err := tea.LogToFile("debug.log", "debug")		//Use  tail -f debug.log to view log while program is running
	// if err != nil {
	// 	fmt.Println("fatal:", err)
	// 	os.Exit(1)
	// }
	// log.Println("Logging...")
	// defer f.Close()

	m := model{
		address: address.New(0, ch_browse, "i=84"),
		data:    data.New(1, ch_read, ch_write, []types.DataPoint{}, types.MyConfig.UpdateRate),
		footer:  footer.New(types.MyConfig.Server.Endpoint),
		overlay: overlay.New(),
	}

	time.Sleep(1000 * time.Millisecond) //wait for OPC UA service connection

	_, _ = tea.NewProgram(m).Run()
}
