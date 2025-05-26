package cmd

import (
	"fmt"
	"os"

	"mark/internal/remote"

	"github.com/spf13/cobra"
)

var addContextItemTextCmd = &cobra.Command{
	Use:   "add-context-item-text <message>",
	Short: "Add a text item to the context",
	Args:  cobra.ExactArgs(1),
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

		err = client.SendMessage("add-context-item-text", args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(addContextItemTextCmd)
}
