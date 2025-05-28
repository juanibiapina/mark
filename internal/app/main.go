package app

import (
	"mark/internal/util"

	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type Main struct {
	contextItemsList *ContextItemsList
	messagesViewport viewport.Model

	hasFocus bool
}

func NewMain() *Main {
	main := &Main{
		contextItemsList: NewContextItemsList(),
	}

	main.Focus()

	return main
}

func (main *Main) Focus() {
	main.hasFocus = true
	main.contextItemsList.Focus()
}

func (main *Main) Blur() {
	main.hasFocus = false
	main.contextItemsList.Blur()
}

func (main *Main) SetSize(width, height int) {
	borderSize := 2 // 2 times the border width

	availableHeight := height - borderSize
	sidebarWidth := width / 3
	messagesWidth := width - sidebarWidth

	main.contextItemsList.SetSize(sidebarWidth-borderSize, availableHeight)

	main.messagesViewport.SetWidth(messagesWidth - borderSize)
	main.messagesViewport.SetHeight(availableHeight)
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
		case "enter":
			var cmd tea.Cmd
			cmd = app.submitMessage()
			cmds = append(cmds, cmd)
		case "shift+j":
			inputHandled = true
			main.messagesViewport.LineDown(1)
		case "shift+k":
			inputHandled = true
			main.messagesViewport.LineUp(1)
		case "ctrl+n":
			inputHandled = true
			app.newSession()
		case "ctrl+c":
			inputHandled = true
			app.agent.Cancel()
		}
	}

	if !inputHandled {
		cmd := main.contextItemsList.Update(app, msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

func (main *Main) View() string {
	sidebar := lipgloss.JoinVertical(lipgloss.Top, main.contextItemsListView())
	mainpane := main.messagesView()
	return lipgloss.JoinHorizontal(lipgloss.Left, sidebar, mainpane)
}

func (main *Main) contextItemsListView() string {
	return util.RenderBorderWithTitle(
		main.contextItemsList.View(),
		main.borderIfFocused(),
		"Context",
		main.panelTitleStyleIfFocused(),
	)
}

func (main *Main) messagesView() string {
	return util.RenderBorderWithTitle(
		main.messagesViewport.View(),
		borderStyle,
		"Messages",
		textStyle,
	)
}

func (main *Main) borderIfFocused() lipgloss.Style {
	if main.hasFocus {
		return focusedBorderStyle
	}
	return borderStyle
}

func (main *Main) panelTitleStyleIfFocused() lipgloss.Style {
	if main.hasFocus {
		return focusedPanelTitleStyle
	}

	return textStyle
}
