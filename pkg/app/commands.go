package app

import (
	"os"
	"os/exec"
	"path"

	"mark/pkg/model"
	"mark/pkg/util"

	tea "github.com/charmbracelet/bubbletea/v2"
)

func (m *App) saveThread() tea.Cmd {
	return func() tea.Msg {
		err := m.db.SaveThread(m.thread)
		if err != nil {
			return errMsg{err}
		}

		return m.loadThreads()()
	}
}

func (m *App) loadSelectedThread() tea.Cmd {
	if len(m.threadListEntries) == 0 {
		return nil
	}

	selectedEntry := m.threadListEntries[m.threadListCursor]

	return func() tea.Msg {
		thread, err := m.db.LoadThread(selectedEntry.ID)
		if err != nil {
			return errMsg{err}
		}

		return threadMsg{thread}
	}
}

func (m *App) loadThreads() tea.Cmd {
	return func() tea.Msg {
		threads, err := m.db.ListThreads()
		if err != nil {
			return errMsg{err}
		}

		return threadEntriesMsg(threads)
	}
}

func (m *App) submitMessage() tea.Cmd {
	m.cancelStreaming()

	v := m.input.Value()
	if v != "" {
		m.thread.AddMessage(model.Message{Role: model.RoleUser, Content: v})
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

func (m *App) deleteSelectedThread() tea.Cmd {
	if len(m.threadListEntries) == 0 {
		return nil
	}

	selectedEntryID := m.threadListEntries[m.threadListCursor].ID

	// Remove the thread from the list of entries
	for i, entry := range m.threadListEntries {
		if entry.ID == selectedEntryID {
			m.threadListEntries = append(m.threadListEntries[:i], m.threadListEntries[i+1:]...)
			break
		}
	}

	// Ensure the cursor is in a valid position
	if len(m.threadListEntries) == 0 {
		m.threadListCursor = 0
	} else {
		m.threadListCursor = util.Clamp(m.threadListCursor, 0, len(m.threadListEntries) - 1)
	}

	m.renderActiveThread()
	m.renderThreadList()

	return func() tea.Msg {
		err := m.db.DeleteThread(selectedEntryID)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func complete(m *App) tea.Cmd {
	return func() tea.Msg {
		err := m.ai.CompleteStreaming(&m.thread, m.stream, m.project)
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

func (m *App) editCommit() (tea.Cmd, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, nil
	}

	tmpdir, err := os.MkdirTemp("", "mark-*")
	if err != nil {
		return nil, err
	}

	tmpFile := path.Join(tmpdir, "mark-pull-request-title")
	err = os.WriteFile(tmpFile, []byte(m.thread.Commit.Description), 0o644)
	if err != nil {
		return nil, err
	}

	c := exec.Command(editor, tmpFile)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		defer os.RemoveAll(tmpdir)

		if err != nil {
			return errMsg{err}
		}

		contents, err := os.ReadFile(tmpFile)
		if err != nil {
			return errMsg{err}
		}

		return commitMsg(string(contents))
	}), nil
}

func (m *App) viewThreadInEditor() (tea.Cmd, error) {
	// build the content
	var content string
	for _, msg := range m.thread.Messages {
		content += "---\n"
		content += msg.Content + "\n"
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, nil
	}

	tmpdir, err := os.MkdirTemp("", "mark-*")
	if err != nil {
		return nil, err
	}

	tmpFile := path.Join(tmpdir, "mark-thread")
	err = os.WriteFile(tmpFile, []byte(content), 0o644)
	if err != nil {
		return nil, err
	}

	c := exec.Command(editor, tmpFile)

	return tea.ExecProcess(c, func(err error) tea.Msg {
		defer os.RemoveAll(tmpdir)

		if err != nil {
			return errMsg{err}
		}

		return nil
	}), nil
}

func (m *App) createCommit() tea.Cmd {
	if m.thread.Commit.Description == "" {
		return nil
	}

	return func() tea.Msg {
		cmd := exec.Command("git", "commit", "-m", m.thread.Commit.Description)

		err := cmd.Run()
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}
