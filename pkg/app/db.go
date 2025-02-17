package app

import (
	"encoding/json"
	"mark/pkg/model"
	"os"
	"path"
)

type Database interface {
	SaveConversation(model.Conversation) error
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

	err = os.WriteFile(path.Join(dir, filename), json, 0644)
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

func (self FilesystemDatabase) ensureDirectory(name string) (string, error) {
	dir := path.Join(self.directory, name)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}
	return dir, nil
}
