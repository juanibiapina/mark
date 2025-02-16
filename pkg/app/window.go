package app

import (
	"mark/pkg/view"
)

func (m *App) renderWindow() string {
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
