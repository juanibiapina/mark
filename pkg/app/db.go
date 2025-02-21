package app

import (
	"encoding/json"
	"os"
	"path"
	"slices"

	"mark/pkg/model"
)

type Database interface {
	SaveThread(model.Thread) error
	LoadThread(string) (model.Thread, error)
	ListThreads() ([]model.ThreadEntry, error)
	DeleteThread(string) error
}

func (self FilesystemDatabase) DeleteThread(id string) error {
	dir, err := self.ensureDirectory("threads")
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

func (self FilesystemDatabase) SaveThread(c model.Thread) error {
	dir, err := self.ensureDirectory("threads")
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

func (self FilesystemDatabase) LoadThread(id string) (model.Thread, error) {
	dir, err := self.ensureDirectory("threads")
	if err != nil {
		return model.Thread{}, err
	}

	filename := id + ".json"

	data, err := os.ReadFile(path.Join(dir, filename))
	if err != nil {
		return model.Thread{}, err
	}

	var c model.Thread
	err = json.Unmarshal(data, &c)
	if err != nil {
		return model.Thread{}, err
	}

	return c, nil
}

func (self FilesystemDatabase) ListThreads() ([]model.ThreadEntry, error) {
	dir, err := self.ensureDirectory("threads")
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	entries := []model.ThreadEntry{}

	for _, file := range files {
		c := model.ThreadEntry{
			ID: file.Name()[:len(file.Name())-5], // remove ".json"
		}

		entries = append(entries, c)
	}

	// reverse the order so the most recent threads are at the top
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
