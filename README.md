OPC UA browser based on the bubbletea TUI framework: https://github.com/charmbracelet/bubbletea

This project is currently in pre-alpha status


Sample configuration file:

config.yaml

server:
  endpoint: "opc.tcp://192.168.100.101:4840"
  policy: "Basic256Sha256"
  mode: "SignAndEncrypt"

authorization:
  certificate: "/home/bostroemc/go/src/github.com/bostroemc/tui/opcua-browser/certificates/rymden-software.certificate.pem" # path to certificate file
  key: "/home/bostroemc/.ssh/private/rymden-software.private.key.pem" # path to private key
  username: "<username>" # username may alse be passed to executable as flag: ./opcua-browser -username="username"
  password: "<password>" # password may also be passed to executable as flag:  ./opcua-browser -password="****" (preferred)

update_rate: 500 # update rate in milliseconds

keybinds:
  # global:
  - action: "quit"
    keys: ["ctrl+q"]
  - action: "toggle_edit_mode"
    keys: ["ctrl+e"]
  - action: "toggle_autoupdate"
    keys: ["ctrl+a"]
  - action: "move_up"
    keys: ["k", "up"]
  - action: "move_down"
    keys: ["j", "down"]
  - action: "push"
    keys: ["p"]
  - action: "toggle_focus"
    keys: ["tab"]
  - action: "delete"
    keys: ["d", "delete"]
  - action: "select"
    keys: ["enter", "right"]
  - action: "back"
    keys: ["u", "left"]



  # shortcuts
  - action: "root"
    keys: ["r"]
    params:
      nodeid: "i=84"
  - action: "user_defined_node"
    keys: ["f"]
    params:
      nodeid: "ns=8;s=plc/app/Application/sym"
