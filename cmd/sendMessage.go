package cmd

import (
	"bufio"
	"fmt"
	"os"

	"mark/pkg/app"

	"github.com/spf13/cobra"
)

// sendMessageCmd subcommand to send messages to mark
var sendMessageCmd = &cobra.Command{
	Use:   "sendMessage",
	Short: "Sends a message to Mark",
	Run: func(cmd *cobra.Command, args []string) {
		// get the current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Open a connection to stdin
		fi, err := os.Stdin.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Check if stdin is a pipe
		if fi.Mode()&os.ModeNamedPipe == 0 {
			fmt.Fprintf(os.Stderr, "Error: stdin is not a pipe\n")
			os.Exit(1)
		}

		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		line := scanner.Text() + "\n"

		// create a new client
		client, err := app.NewClient(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = client.SendMessage(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Message sent")
	},
}

func init() {
	rootCmd.AddCommand(sendMessageCmd)
}
