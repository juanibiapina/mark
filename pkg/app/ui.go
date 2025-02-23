package app

import (
	"mark/pkg/util"

	"github.com/charmbracelet/lipgloss/v2"
)

var (
	textColor              = lipgloss.NoColor{}
	focusColor             = lipgloss.Color("2")

	textStyle              = lipgloss.NewStyle().Foreground(textColor)
	focusedPanelTitleStyle = lipgloss.NewStyle().Foreground(focusColor)

	borderStyle            = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	focusedBorderStyle     = borderStyle.BorderForeground(focusColor)
)

func (m *App) windowView() string {
	return lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarView(), m.mainView())
}

func (m *App) sidebarView() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.inputView(), m.pullRequestView(), m.threadListView())
}

func (m *App) mainView() string {
	return util.RenderBorderWithTitle(
		m.threadView(),
		m.borderIfFocused(FocusedThread),
		"Thread",
		m.panelTitleStyleIfFocused(FocusedThread),
	)
}

func (m *App) inputView() string {
	return util.RenderBorderWithTitle(
		m.input.View(),
		m.borderIfFocused(FocusedInput),
		"Message Assistant",
		m.panelTitleStyleIfFocused(FocusedInput),
	)
}

func (m *App) pullRequestView() string {
	return util.RenderBorderWithTitle(
		m.pullRequestViewport.View(),
		m.borderIfFocused(FocusedPullRequest),
		"Pull Request",
		m.panelTitleStyleIfFocused(FocusedPullRequest),
	)
}

func (m *App) threadListView() string {
	return util.RenderBorderWithTitle(
		m.threadList.View(),
		m.borderIfFocused(FocusedThreadList),
		"Threads",
		m.panelTitleStyleIfFocused(FocusedThreadList),
	)
}

func (m *App) panelTitleStyleIfFocused(focused Focused) lipgloss.Style {
	if m.focused == focused {
		return focusedPanelTitleStyle
	}
	return textStyle
}

func (m *App) borderIfFocused(focused Focused) lipgloss.Style {
	if m.focused == focused {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) threadView() string {
	return m.threadViewport.View()
}

func (m *App) handleWindowSize(width, height int) {
	borderSize := 2 // 2 times the border width

	m.mainPanelWidth = int(float64(width) * ratio)
	m.mainPanelHeight = height
	m.sideBarWidth = width - m.mainPanelWidth

	m.input.SetWidth(m.sideBarWidth - borderSize)
	m.input.SetHeight(inputHeight - borderSize)
	rest := height - inputHeight
	half := rest / 2
	m.pullRequestViewport.SetWidth(m.sideBarWidth - borderSize)
	m.pullRequestViewport.SetHeight(half - borderSize)
	m.threadList.SetWidth(m.sideBarWidth - borderSize)
	m.threadList.SetHeight(half - borderSize)
	highlightedEntryStyle = highlightedEntryStyle.Width(m.sideBarWidth - borderSize)

	m.threadViewport.SetWidth(m.mainPanelWidth - 2)   // 2 is the border width
	m.threadViewport.SetHeight(m.mainPanelHeight - 2) // 2 is the border width

	if !m.uiReady {
		m.uiReady = true
	}
}
