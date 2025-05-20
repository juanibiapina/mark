package app

import (
	"mark/internal/util"

	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type Focused int

const (
	FocusedInput Focused = iota
	FocusedContextItemsList
	FocusedMessages
	FocusedEndMarker // used to determine the number of focusable items for cycling
)

type Main struct {
	input            textarea.Model
	contextItemsList *ContextItemsList
	messagesViewport viewport.Model

	focused Focused
}

func NewMain() *Main {
	input := textarea.New()
	input.Focus()
	input.CharLimit = 0 // no character limit
	input.MaxHeight = 0 // no max height
	input.Prompt = ""
	input.Styles.Focused.CursorLine = lipgloss.NewStyle() // Remove cursor line styling
	input.ShowLineNumbers = false
	input.KeyMap.InsertNewline.SetEnabled(false)

	main := &Main{
		input:            input,
		contextItemsList: NewContextItemsList(),
	}

	return main
}

func (main *Main) SetSize(width, height int) {
	borderSize := 2 // 2 times the border width
	inputHeight := 5

	availableHeight := height - borderSize

	main.input.SetWidth(width - borderSize)
	main.input.SetHeight(inputHeight - borderSize)

	sidebarWidth := width / 3
	messagesWidth := width - sidebarWidth

	main.contextItemsList.SetSize(sidebarWidth-borderSize, availableHeight-inputHeight)

	main.messagesViewport.SetWidth(messagesWidth - borderSize)
	main.messagesViewport.SetHeight(availableHeight - inputHeight)
}

func (main *Main) Update(app *App, msg tea.Msg) tea.Cmd {
	var inputHandled bool
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			inputHandled = true
			cmds = append(cmds, tea.Quit)
		case "tab":
			inputHandled = true
			main.focusNext()
		case "shift+tab":
			inputHandled = true
			main.focusPrev()
		case "enter":
			inputHandled = true
			cmd := app.submitMessage()
			cmds = append(cmds, cmd)
		case "ctrl+a":
			inputHandled = true
			app.showAddContextDialog()
		case "ctrl+n":
			inputHandled = true
			app.newSession()
		case "ctrl+c":
			inputHandled = true
			app.agent.Cancel()
		}
	}

	if !inputHandled {
		if main.focused == FocusedInput {
			cmd := main.processInputView(msg)
			cmds = append(cmds, cmd)
		}

		if main.focused == FocusedContextItemsList {
			cmd := main.contextItemsList.Update(app, msg)
			cmds = append(cmds, cmd)
		}

		if main.focused == FocusedMessages {
			cmd := main.processMessagesView(app, msg)
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

func (main *Main) View() string {
	top := lipgloss.JoinHorizontal(lipgloss.Left, main.contextItemsListView(), main.messagesView())
	bottom := main.inputView()
	return lipgloss.JoinVertical(lipgloss.Top, top, bottom)
}

func (main *Main) inputView() string {
	return util.RenderBorderWithTitle(
		main.input.View(),
		main.borderIfFocused(FocusedInput),
		"Message Assistant",
		main.panelTitleStyleIfFocused(FocusedInput),
	)
}

func (main *Main) contextItemsListView() string {
	return util.RenderBorderWithTitle(
		main.contextItemsList.View(),
		main.borderIfFocused(FocusedContextItemsList),
		"Context",
		main.panelTitleStyleIfFocused(FocusedContextItemsList),
	)
}

func (main *Main) messagesView() string {
	return util.RenderBorderWithTitle(
		main.messagesViewport.View(),
		main.borderIfFocused(FocusedMessages),
		"Messages",
		main.panelTitleStyleIfFocused(FocusedMessages),
	)
}

func (main *Main) focusNext() {
	main.focused += 1
	if main.focused == FocusedEndMarker {
		main.focused = 0
	}

	if main.focused == FocusedContextItemsList {
		main.contextItemsList.Focus()
	} else {
		main.contextItemsList.Blur()
	}
}

func (main *Main) focusPrev() {
	main.focused -= 1
	if main.focused < 0 {
		main.focused = FocusedEndMarker - 1
	}

	if main.focused == FocusedContextItemsList {
		main.contextItemsList.Focus()
	} else {
		main.contextItemsList.Blur()
	}
}

func (main *Main) borderIfFocused(focused Focused) lipgloss.Style {
	if main.focused == focused {
		return focusedBorderStyle
	}
	return borderStyle
}

func (main *Main) panelTitleStyleIfFocused(focused Focused) lipgloss.Style {
	if main.focused == focused {
		return focusedPanelTitleStyle
	}
	return textStyle
}

func (main *Main) processInputView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	main.input, cmd = main.input.Update(msg)
	return cmd
}

func (main *Main) processMessagesView(app *App, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "e":
			cmd, err := app.viewMessagesInEditor()
			if err != nil {
				app.err = err
				return tea.Quit
			}

			return cmd
		default:
			var cmd tea.Cmd
			main.messagesViewport, cmd = main.messagesViewport.Update(msg)
			return cmd
		}
	}

	return nil
}
