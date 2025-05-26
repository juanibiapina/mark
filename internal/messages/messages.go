package messages

import (
	"fmt"

	"mark/internal/app"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Message struct {
	Use      string
	Short    string
	NumArgs  int
	ToTeaMsg func(args []string) tea.Msg
}

var Msgs map[string]Message = map[string]Message{
	"add-context-item-text": {
		Use:     "add-context-item-text <message>",
		Short:   "Add a text item to the context",
		NumArgs: 1,
		ToTeaMsg: func(args []string) tea.Msg {
			return app.AddContextItemTextMsg(args[0])
		},
	},
	"add-context-item-file": {
		Use:     "add-context-item-file <path>",
		Short:   "Add a file item to the context",
		NumArgs: 1,
		ToTeaMsg: func(args []string) tea.Msg {
			return app.AddContextItemFileMsg(args[0])
		},
	},
	"run": {
		Use:     "run",
		Short:   "Run the agent",
		NumArgs: 0,
		ToTeaMsg: func(args []string) tea.Msg {
			return app.RunMsg{}
		},
	},
}

func ToTeaMsg(msg string, args []string) tea.Msg {
	message, ok := Msgs[msg]
	if ok {
		return message.ToTeaMsg(args)
	}

	return app.ErrMsg{Err: fmt.Errorf("unknown message: %s", msg)}
}
