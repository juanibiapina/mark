package app

import (
	"slices"
	"strings"

	"mark/pkg/model"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/samber/lo"
)

type PromptList struct {
	viewport viewport.Model

	selected *string
	prompts  []model.Prompt
}

// startinterface: tea.Model

func (i PromptList) Init() (PromptList, tea.Cmd) {
	return i, nil
}

func (i PromptList) Update(msg tea.Msg) (PromptList, tea.Cmd) {
	return i, nil
}

func (i PromptList) View() string {
	return ""
}

// endinterface: tea.Model

// startinterface: Container

func (i PromptList) Render(width, height int) string {
	i.viewport.Width = width
	i.viewport.Height = height

	i.renderPrompts()

	return i.viewport.View()
}

// endinterface: Container

func (i *PromptList) renderPrompts() {
	names := lo.Map(i.prompts, func(p model.Prompt, _ int) string {
		return p.Name()
	})

	slices.Sort(names)

	content := strings.Join(names, "\n")

	i.viewport.SetContent(content)
}
