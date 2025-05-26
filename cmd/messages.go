package cmd

import (
	"fmt"
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
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				client, err := remote.NewClient(cwd)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				err = client.SendMessage(name, args)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
