package app

import (
	"fmt"
	"io"
	"strings"

	"mark/internal/llm"

	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().Background(lipgloss.Color("4"))
)

type contextItemDelegate struct {
	l *ContextItemsList
}

func (d *contextItemDelegate) Height() int                             { return 1 }
func (d *contextItemDelegate) Spacing() int                            { return 0 }
func (d *contextItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d *contextItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(llm.ContextItem)
	if !ok {
		return
	}

	maxWidth := d.l.Width()

	str := string(i)
	str = ansi.Truncate(str, maxWidth-2, "...") // - 2 for padding

	fn := itemStyle.Width(maxWidth).Render
	if d.l.IsFocused() {
		if index == m.Index() {
			fn = func(s ...string) string {
				return selectedItemStyle.Width(maxWidth).Render("> " + strings.Join(s, " "))
			}
		}
	}

	fmt.Fprint(w, fn(str))
}

func (d *contextItemDelegate) SetContextItemsList(l *ContextItemsList) {
	d.l = l
}

type item string

func (i item) FilterValue() string { return "" }

type ContextItemsList struct {
	focused bool
	model   list.Model
}

func NewContextItemsList() *ContextItemsList {
	delegate := &contextItemDelegate{}

	model := list.New([]list.Item{}, delegate, 0, 0)
	model.DisableQuitKeybindings()
	model.SetShowStatusBar(false)
	model.SetFilteringEnabled(false)
	model.SetShowHelp(false)
	model.SetShowTitle(false)
	model.SetStatusBarItemName("Context", "Context")

	contextItemsList := &ContextItemsList{
		model: model,
	}

	delegate.SetContextItemsList(contextItemsList)

	return contextItemsList
}

func (l *ContextItemsList) SetSize(width, height int) {
	l.model.SetWidth(width) // this only sets the max width of items
	l.model.SetHeight(height)
}

func (l *ContextItemsList) Width() int {
	return l.model.Width()
}

func (l *ContextItemsList) Update(app *App, msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	l.model, cmd = l.model.Update(msg)
	return cmd
}

func (l *ContextItemsList) View() string {
	return lipgloss.NewStyle().Width(l.model.Width()).Render(l.model.View())
}

func (l *ContextItemsList) Focus() {
	l.focused = true
}

func (l *ContextItemsList) Blur() {
	l.focused = false
}

func (l *ContextItemsList) IsFocused() bool {
	return l.focused
}

func (l *ContextItemsList) SetItemsFromSessionContext(context []llm.ContextItem) {
	items := make([]list.Item, len(context))
	for i, ctxItem := range context {
		items[i] = ctxItem
	}
	l.model.SetItems(items)
}
