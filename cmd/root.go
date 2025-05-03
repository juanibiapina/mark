// Package cmd provides the command line interface for the Mark TUI Assistant.
// It uses cobra to define and manage commands.
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"mark/internal/app"

	"github.com/spf13/cobra"
)

// rootCmd runs when the cli is invoked without any subcommands.
// It initializes and runs the Program, printing errors to stderr and exiting
// with a non-zero status code when an error occurs.
var rootCmd = &cobra.Command{
	Use:   "mark",
	Short: "Mark TUI Assistant",
	Run: func(cmd *cobra.Command, args []string) {
		setupLogging()

		program, err := app.NewProgram()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = program.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
			os.Exit(1)
		}
	},
}

// Execute the root command.
// This is called by main.main().
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// setupLogging sets up logging to a file if the DEBUG environment variable is set.
func setupLogging() {
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

// init defines flags and configuration settings.
func init() {
	// Cobra supports persistent flags, which, if defined here,
	// will be global for the application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mark.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
