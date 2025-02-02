package main

import (
	"fmt"
	"os"

	"mark/pkg/app"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mark",
		Short: "Mark TUI Assistant",
		Run: func(cmd *cobra.Command, args []string) {
			if len(os.Getenv("DEBUG")) > 0 {
				f, err := tea.LogToFile("debug.log", "debug")
				if err != nil {
					fmt.Println("fatal:", err)
					os.Exit(1)
				}
				defer f.Close()
			}

			app, err := app.MakeApp()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithKeyboardEnhancements())
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
