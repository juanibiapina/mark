package app

import (
	"os"
	"path"

	"mark/pkg/model"

	"github.com/charmbracelet/bubbles/v2/cursor"
	"github.com/charmbracelet/bubbles/v2/textarea"
	"github.com/charmbracelet/lipgloss/v2"
)

func (m *App) initInput() {
	m.input = textarea.New()
	m.input.Focus() // focus is actually handled by the app

	m.input.Cursor.SetMode(cursor.CursorStatic)

	m.input.Prompt = ""

	m.input.SetHeight(3)

	// Remove cursor line styling
	m.input.Styles.Focused.CursorLine = lipgloss.NewStyle()

	m.input.ShowLineNumbers = false

	m.input.KeyMap.InsertNewline.SetEnabled(false)
}

func (m *App) initPrompts() error {
	dir := m.config.promptsDir

	prompts := []model.Prompt{}

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
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

	m.prompts = prompts

	return nil
}
