package logging

import "log/slog"

func NewLogger(component string) *slog.Logger {
	logger := slog.Default().With("component", component)

	return logger
}
