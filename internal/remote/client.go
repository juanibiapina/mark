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

type Request struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

func NewClient(cwd string) (*Client, error) {
	client := Client{
		socketPath: path.Join(cwd, ".local", "share", "mark", "socket"),
	}
	return &client, nil
}

func (client *Client) SendRequest(req Request) error {
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

	// Convert the message to JSON format
	data, err := json.Marshal(&req)
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
