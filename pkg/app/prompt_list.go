package app

import (
	"mark/pkg/model"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
)

type PromptList struct {
	viewport viewport.Model

	prompts       []model.Prompt
	selectedIndex int
	focus         bool
}

func (i *PromptList) SelectedIndex() int {
	return i.selectedIndex
}

// startinterface: tea.Model

func (i PromptList) Init() (PromptList, tea.Cmd) {
	return i, nil
}

func (i PromptList) Update(msg tea.Msg) (PromptList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j", "down":
			i.incSelectedIndex()

		case "k", "up":
			i.decSelectedIndex()
		}
	}

	return i, nil
}

func (i PromptList) View() string {
	return ""
}

// endinterface: tea.Model

func (i *PromptList) incSelectedIndex() {
	i.selectedIndex++
	if i.selectedIndex >= len(i.prompts) {
		i.selectedIndex = 0
	}
}

func (i *PromptList) decSelectedIndex() {
	i.selectedIndex--
	if i.selectedIndex < 0 {
		i.selectedIndex = len(i.prompts) - 1
	}
}

// startinterface: Container

func (i PromptList) Render(width, height int) string {
	i.viewport.Width = width
	i.viewport.Height = height

	i.renderPrompts()

	return i.viewport.View()
}

// endinterface: Container

func (pl *PromptList) Focus() {
	pl.focus = true
}

func (pl *PromptList) Blur() {
	pl.focus = false
}

func (pl *PromptList) renderPrompts() {
	var content string

	for index, prompt := range pl.prompts {
		name := prompt.Name()

		var prefix string
		if index == pl.selectedIndex {
			prefix = " "
		} else {
			prefix = "  "
		}

		content += prefix + name + "\n"

	}

	pl.viewport.SetContent(content)
}
