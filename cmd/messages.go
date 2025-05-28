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
	for command, msg := range messages.Msgs {
		// Variables that capture flag values
		var useStdin bool

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

				// Create a new remote client
				client, err := remote.NewClient(cwd)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				var stdin string
				if useStdin {
					stdinData, err := io.ReadAll(os.Stdin)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
						os.Exit(1)
					}
					stdin = string(stdinData)
				}

				// Send the message using the client
				err = client.SendRequest(remote.Request{Command: command, Args: args, Stdin: stdin})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			},
		}

		if msg.StdinFlagEnabled {
			cmd.Flags().BoolVar(&useStdin, "stdin", false, "Also read stdin")
		}
		rootCmd.AddCommand(cmd)
	}
}
