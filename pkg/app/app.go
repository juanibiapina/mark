package app

import (
	"log"
	"os"
	"path"
	"reflect"

	"mark/pkg/model"
	"mark/pkg/openai"
	"mark/pkg/view"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

type Focused int

const (
	FocusedInput Focused = iota
	FocusedPromptList
	FocusedConversation
	FocusedEndMarker // used to determine the number of focusable items for cycling
)

var (
	borderStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	focusedBorderStyle = borderStyle.BorderForeground(lipgloss.Color("2"))
)

type App struct {
	// app state
	uiReady       bool
	width, height int
	focused       Focused

	// models
	prompts      []model.Prompt
	conversation model.Conversation

	// streaming
	streaming      bool
	stream         *model.StreamingMessage
	partialMessage string

	// view models
	conversationView Conversation
	input            Input
	promptListView   PromptList

	// llm
	ai *openai.OpenAI

	// error
	err error
}

func MakeApp(c Config) (App, error) {
	// Load prompts from files
	prompts, err := loadPrompts(c.promptsDir)
	if err != nil {
		return App{}, err
	}

	app := App{
		focused: FocusedInput,
		input:   MakeInput(),
		promptListView: PromptList{
			prompts: prompts,
		},
		ai:      openai.NewOpenAIClient(),
		prompts: prompts,
	}

	// start a new conversation
	app.newConversation()

	return app, nil
}

func (m App) Err() error {
	return m.err
}

// Init Init is required by the bubbletea interface. It is called once when the
// program starts. It is used to send an initial command to the update function.
func (m App) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(os.Getenv("DEBUG")) > 0 {
		log.Print("msg: ", reflect.TypeOf(msg), msg)
	}

	var cmds []tea.Cmd
	var inputHandled bool // whether the key event was handled and shouldn't be passed to the input view

	switch msg := msg.(type) {

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.uiReady {
			m.uiReady = true
		}

	case partialMessage:
		// Ignore message if streaming has been cancelled
		if !m.streaming {
			return m, nil
		}

		m.partialMessage += string(msg)

		cmds = append(cmds, processStream(&m))

	case replyMessage:
		// Ignore message if streaming has been cancelled
		if !m.streaming {
			return m, nil
		}

		m.streaming = false
		m.partialMessage = ""
		m.conversation.AddMessage(model.Message{Role: model.RoleAssistant, Content: string(msg)})

	case tea.KeyPressMsg:
		switch msg.String() {

		case "esc":
			return m, tea.Quit

		case "tab":
			m.focusNext()
			inputHandled = true

		case "shift+tab":
			m.focusPrev()
			inputHandled = true

		case "enter":
			if m.focused == FocusedInput {
				cmd := m.submitMessage()
				cmds = append(cmds, cmd)
				inputHandled = true
			}

			if m.focused == FocusedPromptList {
				prompt := m.promptListView.SelectedIndex()
				m.conversation.SetPrompt(m.prompts[prompt])
				inputHandled = true
			}

		case "ctrl+n":
			m.newConversation()
			inputHandled = true

		case "ctrl+c":
			m.cancelStreaming()
			inputHandled = true

		}
	}

	if m.uiReady {
		if m.focused == FocusedInput && !inputHandled {
			cmd := m.processInputView(msg)
			cmds = append(cmds, cmd)
		}

		if m.focused == FocusedPromptList {
			cmd := m.processPromptListView(msg)
			cmds = append(cmds, cmd)
		}

		m.conversationView.Set(&m.conversation, m.streaming, m.partialMessage)
	}

	return m, tea.Batch(cmds...)
}

func (m App) View() string {
	if !m.uiReady {
		return "Initializing..."
	}

	main := view.Main{
		Left: view.Sidebar{
			Input:   view.NewPane(m.input, m.borderInput(), "Message Assistant"),
			Prompts: view.NewPane(m.promptListView, m.borderPromptList(), "Prompts"),
		},
		Right: view.NewPane(m.conversationView, m.borderConversation(), "Conversation"),
		Ratio: 0.67,
	}

	return main.Render(m.width, m.height)
}

func (m *App) focusNext() {
	m.focused += 1
	if m.focused == FocusedEndMarker {
		m.focused = 0
	}

	m.updateFocus()
}

func (m *App) focusPrev() {
	m.focused -= 1
	if m.focused < 0 {
		m.focused = FocusedEndMarker - 1
	}

	m.updateFocus()
}

// updateFocus update focus on each individual view based on the current focus
func (m *App) updateFocus() {
	switch m.focused {
	case FocusedInput:
		m.promptListView.Blur()
	case FocusedPromptList:
		m.promptListView.Focus()
	case FocusedConversation:
		m.promptListView.Blur()
	}
}

func (m *App) borderInput() lipgloss.Style {
	if m.focused == FocusedInput {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderPromptList() lipgloss.Style {
	if m.focused == FocusedPromptList {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) borderConversation() lipgloss.Style {
	if m.focused == FocusedConversation {
		return focusedBorderStyle
	}
	return borderStyle
}

func (m *App) startStreaming() {
	m.stream = model.NewStreamingMessage()
	m.streaming = true
}

func (m *App) cancelStreaming() {
	if m.stream != nil {
		m.stream.Cancel()
	}
	m.stream = nil

	m.streaming = false

	// Add the partial message to the chat history
	m.conversation.AddMessage(model.Message{Role: model.RoleAssistant, Content: m.partialMessage})

	m.partialMessage = ""
}

func (m *App) processInputView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *App) processPromptListView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.promptListView, cmd = m.promptListView.Update(msg)
	return cmd
}

// newConversation starts a new conversation without a prompt
func (m *App) newConversation() {
	m.cancelStreaming()

	m.conversation = model.MakeConversation()

	m.input.Reset()

	m.focused = FocusedInput
}

func (m *App) submitMessage() tea.Cmd {
	m.cancelStreaming()

	v := m.input.Value()
	if v != "" {
		m.conversation.AddMessage(model.Message{Role: model.RoleUser, Content: v})
		m.input.Reset()
	}

	// maybe update the prompt here

	m.startStreaming()

	cmds := []tea.Cmd{
		complete(m),      // call completions API
		processStream(m), // start receiving partial message
	}
	return tea.Batch(cmds...)
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		//rs := llm.ResponseSchema{
		//	Name: "replace_line",
		//	Description: "Replace a line in a file",
		//	Schema: ReplaceLineResponseSchema,
		//}

		// replaceLine := ReplaceLine{}

		err := m.ai.CompleteStreaming(&m.conversation, m.stream)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func processStream(m *App) tea.Cmd {
	return func() tea.Msg {
		select {
		case v := <-m.stream.Reply:
			return replyMessage(v)
		case v := <-m.stream.Chunks:
			return partialMessage(v)
		}
	}
}

func loadPrompts(dir string) ([]model.Prompt, error) {
	prompts := []model.Prompt{}

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return prompts, nil
		}

		return nil, err
	}

	// add each .md file as a prompt
	for _, file := range files {
		if file.IsDir() {
			// skip directories
			continue
		}

		filename := file.Name()
		if filename[len(filename)-3:] != ".md" {
			// skip non-markdown files
			continue
		}

		prompt := model.MakePromptFromFile(filename, path.Join(dir, filename))
		prompts = append(prompts, prompt)
	}

	return prompts, nil
}
