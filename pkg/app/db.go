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
	json, err := json.Marshal(c)
	if err != nil {
		return err
	}

	filename := c.ID + ".json"

	err = os.WriteFile(path.Join(self.directory, filename), json, 0644)
	if err != nil {
		return err
	}

	return nil
}
