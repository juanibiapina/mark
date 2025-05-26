package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path"
)

type Client struct {
	socketPath string
}

type ClientRequest struct {
	Message string   `json:"message"`
	Args    []string `json:"args,omitempty"`
}

func NewClient(cwd string) (*Client, error) {
	client := Client{
		socketPath: path.Join(cwd, ".local", "share", "mark", "socket"),
	}
	return &client, nil
}

func (client *Client) SendMessage(message string, args []string) error {
	// Check if the socket exists
	_, err := os.Stat(client.socketPath)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("Couldn't find socket path: %s", client.socketPath)
	}

	// open the socket connection
	conn, err := client.openSocketConnection(client.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Prepare the message to send
	cr := ClientRequest{
		Message: message,
		Args:    args,
	}
	// Convert the message to JSON format
	data, err := json.Marshal(&cr)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write the message to the socket
	_, err = conn.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (self *Client) openSocketConnection(socketPath string) (net.Conn, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
