package app

import (
	"encoding/json"
	"os"
	"path"
	"slices"

	"mark/pkg/model"
)

type Database interface {
	SaveConversation(model.Conversation) error
	LoadConversation(string) (model.Conversation, error)
	ListConversations() ([]model.ConversationEntry, error)
	DeleteConversation(string) error
}

func (self FilesystemDatabase) DeleteConversation(id string) error {
	dir, err := self.ensureDirectory("conversations")
	if err != nil {
		return err
	}

	filename := id + ".json"
	filePath := path.Join(dir, filename)

	err = os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}

type FilesystemDatabase struct {
	directory string
}

func MakeFilesystemDatabase(directory string) FilesystemDatabase {
	return FilesystemDatabase{directory: directory}
}

func (self FilesystemDatabase) SaveConversation(c model.Conversation) error {
	dir, err := self.ensureDirectory("conversations")
	if err != nil {
		return err
	}

	json, err := json.Marshal(c)
	if err != nil {
		return err
	}

	filename := c.ID + ".json"

	err = os.WriteFile(path.Join(dir, filename), json, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func (self FilesystemDatabase) LoadConversation(id string) (model.Conversation, error) {
	dir, err := self.ensureDirectory("conversations")
	if err != nil {
		return model.Conversation{}, err
	}

	filename := id + ".json"

	data, err := os.ReadFile(path.Join(dir, filename))
	if err != nil {
		return model.Conversation{}, err
	}

	var c model.Conversation
	err = json.Unmarshal(data, &c)
	if err != nil {
		return model.Conversation{}, err
	}

	return c, nil
}

func (self FilesystemDatabase) ListConversations() ([]model.ConversationEntry, error) {
	dir, err := self.ensureDirectory("conversations")
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	entries := []model.ConversationEntry{}

	for _, file := range files {
		c := model.ConversationEntry{
			ID: file.Name()[:len(file.Name())-5], // remove ".json"
		}

		entries = append(entries, c)
	}

	// reverse the order so the most recent conversations are at the top
	slices.Reverse(entries)

	return entries, nil
}

func (self FilesystemDatabase) ensureDirectory(name string) (string, error) {
	dir := path.Join(self.directory, name)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return "", err
	}
	return dir, nil
}
