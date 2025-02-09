package app

import (
	"path"

	"github.com/adrg/xdg"
)

type Config struct {
	promptsDir string
}

func NewConfig() Config {
	promptsDir := path.Join(xdg.ConfigHome, "mark/prompts")

	return Config{
		promptsDir: promptsDir,
	}
}
