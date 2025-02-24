package app

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path"
)

type Client struct {
	cwd string
}

func NewClient(cwd string) (*Client, error) {
	client := Client{
		cwd: cwd,
	}
	return &client, nil
}

func (self *Client) SendMessage(message string) error {
	// Get the path to the socket
	socketPath := self.getSocketPath()

	// Check if the socket exists
	_, err := os.Stat(socketPath)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("Couldn't find socket path: %s", socketPath)
	}

	// Create a JSON message
	jsonMessage := fmt.Sprintf(`{"message": "%s"}`, message)


	conn, err := self.openSocketConnection(socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Write the JSON message to the socket
	_, err = conn.Write([]byte(jsonMessage))
	if err != nil {
		return err
	}

	return nil
}

func (self *Client) getSocketPath() string {
	return path.Join(self.cwd, ".mark", "socket")
}

func (self *Client) openSocketConnection(socketPath string) (net.Conn, error) {
	conn, err := net.Dial("unix", socketPath)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
