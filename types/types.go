package types

import (
	"log"
	"os"
	"strconv"

	"github.com/gopcua/opcua/ua"
	"gopkg.in/yaml.v3"
)

var KeyActions map[string]KeyAction
var MyConfig Config

type OpcUaBrowserData struct {
	Node     *ua.NodeID //NodeID of the parent
	Parent   Node
	Children []Node
}

type OpcUaReadData struct {
	Data  []DataPoint //List of nodes to be read
	Count int         //count is used to verify that that underlying model has not changed during the async operation
}

type DataPoint struct {
	Enable  bool // TODO: EnableValues
	Block   bool // TODO: Block updateValuesrouter -- to be used if valValuescurrently being edited in UI
	Id      uint32
	Type    string
	Value   any
	Pending any
	Node    string
}

func (d DataPoint) String() string {
	s := ""
	switch d.Value.(type) {
	case int32:
		s = strconv.FormatInt(int64(d.Value.(int32)), 10)

	case int64:
		s = strconv.FormatInt(d.Value.(int64), 10)

	case float32:
		s = strconv.FormatFloat(float64(d.Value.(float32)), 'f', 3, 32)

	case float64:
		s = strconv.FormatFloat(d.Value.(float64), 'f', 3, 64)

	case bool:
		s = strconv.FormatBool(d.Value.(bool))

	case string:
		s = d.Value.(string)
	default:
		s = "Type not found"
	}

	return s
}

func (d *DataPoint) SetValue(s string) (err error) {

	switch d.Value.(type) {
	case int64:
		var temp int64
		temp, err = strconv.ParseInt(s, 10, 64)
		d.Value = temp

	case float64:
		var temp float64
		temp, err = strconv.ParseFloat(s, 64)
		d.Value = temp

	case float32:
		var temp float64
		temp, err = strconv.ParseFloat(s, 32)
		d.Pending = float32(temp)

	case bool:
		var temp bool
		temp, err = strconv.ParseBool(s)
		d.Value = temp

	case string:
		d.Value = s
	}

	return err
}

func (d *DataPoint) SetPending(s string) (err error) {

	switch d.Value.(type) {
	case int64:
		var temp int64
		temp, err = strconv.ParseInt(s, 10, 64)
		d.Pending = temp

	case float64:
		var temp float64
		temp, err = strconv.ParseFloat(s, 64)
		d.Pending = temp

	case float32:
		var temp float64
		temp, err = strconv.ParseFloat(s, 32)
		d.Pending = float32(temp)

	case bool:
		var temp bool
		temp, err = strconv.ParseBool(s)
		d.Pending = temp

	case string:
		d.Pending = s
	}

	return err
}

type Node struct {
	NodeID      *ua.NodeID
	NodeClass   ua.NodeClass
	BrowseName  string
	Description string
	AccessLevel ua.AccessLevelType
	Path        string
	DataType    string
	Writable    bool
	Unit        string
	Scale       string
	Min         string
	Max         string
	Value       any
}

type Server struct {
	Endpoint string `yaml:"endpoint"`
	Policy   string `yaml:"policy"`
	Mode     string `yaml:"mode"`
}
type Authorization struct {
	Certificate string `yaml:"certificate"`
	Key         string `yaml:"key"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
}

//	type Keybinds struct {
//		Global     []keybind `yaml:"global"`
//		Navigation []keybind `yaml:"navigation"`
//		List       []keybind `yaml:"list"`
//	}
type Config struct {
	Server        Server        `yaml:"server"`
	Authorization Authorization `yaml:"authorization"`
	UpdateRate    int           `yaml:"update_rate"`
	Keybinds      []Keybind     `yaml:"keybinds,omitempty"`
}

type Keybind struct {
	Action string   `yaml:"action"`
	Keys   []string `yaml:"keys"`
	Params *Params  `yaml:"params,omitempty"`
}

type Params struct {
	NodeId *string `yaml:"nodeid,omitempty"` //add additional params as required
}
type KeyAction struct { //used to create map: keyActions map[string]types.KeyAction
	Action string
	Params *Params
}

// type KeyAction struct {
// 	Action func(params ...*Params) (tea.Model, tea.Cmd)
// 	Params *Params
// }

func (c *Config) Init() {
	home, _ := os.UserHomeDir()
	f, _ := os.ReadFile(home + "/.config/opcua-browser/config.yaml")

	if err := yaml.Unmarshal(f, c); err != nil {
		log.Fatal(err)
	}

	KeyActions = make(map[string]KeyAction)
	for _, b := range c.Keybinds {
		for _, k := range b.Keys {
			KeyActions[k] = KeyAction{Action: b.Action, Params: b.Params}
		}
	}
}
