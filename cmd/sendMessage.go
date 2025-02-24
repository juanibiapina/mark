package cmd

import (
	"fmt"
	"os"

	"mark/pkg/app"

	"github.com/spf13/cobra"
)

// sendMessageCmd subcommand to send messages to mark
var sendMessageCmd = &cobra.Command{
	Use:   "sendMessage",
	Short: "Sends a message to Mark",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client, err := app.NewClient(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = client.SendMessage(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(sendMessageCmd)
}
