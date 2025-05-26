// Package cmd provides the command line interface for the Mark TUI Assistant.
// It uses cobra to define and manage commands.
package cmd

import (
	"fmt"
	"os"

	"mark/internal/logging"
	"mark/internal/program"

	"github.com/spf13/cobra"
)

// rootCmd runs when the cli is invoked without any subcommands.
// It initializes and runs the Program, printing errors to stderr and exiting
// with a non-zero status code when an error occurs.
var rootCmd = &cobra.Command{
	Use:   "mark",
	Short: "Mark TUI Assistant",
	Run: func(cmd *cobra.Command, args []string) {
		logging.Setup()

		program, err := program.NewProgram()
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

// init defines flags and configuration settings.
func init() {
	// Cobra supports persistent flags, which, if defined here,
	// will be global for the application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mark.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
