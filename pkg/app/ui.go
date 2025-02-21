package app

import (
	"mark/pkg/util"

	"github.com/charmbracelet/lipgloss/v2"
)

func (m *App) windowView() string {
	return lipgloss.JoinHorizontal(lipgloss.Top, m.sidebarView(), m.mainView())
}

func (m *App) sidebarView() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.inputView(), m.threadListView())
}

func (m *App) mainView() string {
	return util.RenderBorderWithTitle(
		m.threadView(),
		m.borderThread(),
		"Thread",
	)
}

func (m *App) inputView() string {
	return util.RenderBorderWithTitle(
		m.input.View(),
		m.borderInput(),
		"Message Assistant",
	)
}

func (m *App) threadListView() string {
	return util.RenderBorderWithTitle(
		m.threadList.View(),
		m.borderThreadList(),
		"Threads",
	)
}

func (m *App) threadView() string {
	return m.threadViewport.View()
}
