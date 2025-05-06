package logging

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
)

func NewLogger(component string) *slog.Logger {
	logger := slog.Default().With("component", component)

	return logger
}

func Setup() {
	if len(os.Getenv("DEBUG")) > 0 {
		file, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
			os.Exit(1)
		}
		log.SetOutput(file)
	} else {
		log.SetOutput(io.Discard)
	}
}
