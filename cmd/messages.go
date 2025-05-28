package cmd

import (
	"fmt"
	"io"
	"os"

	"mark/internal/messages"
	"mark/internal/remote"

	"github.com/spf13/cobra"
)

func init() {
	for name, msg := range messages.Msgs {
		cmd := &cobra.Command{
			Use:   msg.Use,
			Short: msg.Short,
			Args:  cobra.ExactArgs(msg.NumArgs),
			Run: func(cmd *cobra.Command, args []string) {
				// Get the current working directory
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				// Read stdin
				stdinBytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
					os.Exit(1)
				}

				// Create a new remote client
				client, err := remote.NewClient(cwd)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				// Send the message
				req := remote.Request{
					Command: name,
					Args:    args,
					Stdin:   string(stdinBytes),
				}
				if err := client.SendRequest(req); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
