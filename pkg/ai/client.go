package ai

import (
	"context"

	"github.com/openai/openai-go"
)

type Role int

const (
    User Role = iota
    Assistant
)

type Message struct {
	Role    Role
	Content string
}

type Client struct {
	client   *openai.Client
}

func NewClient() *Client {
	return &Client{
		client:   openai.NewClient(),
	}
}

// Complete sends a message to the AI model and returns the response.
func (c *Client) Complete(ctx context.Context, messages []Message) (string, error) {
	// Convert messages to OpenAI format
	var chatMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range messages {
		if msg.Role == User {
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		} else {
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		}
	}

	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(chatMessages),
		Seed:     openai.Int(1),
		Model:    openai.F(openai.ChatModelGPT4o),
	})

	if err != nil {
		return "", err
	}

	// Return the response
	response := completion.Choices[0].Message.Content

	return response, nil
}
