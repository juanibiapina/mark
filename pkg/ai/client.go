package ai

import (
	"context"

	"github.com/openai/openai-go"
)

type Client struct {
	client *openai.Client
}

func NewClient() *Client {
	return &Client{
		client: openai.NewClient(),
	}
}

// SendMessage sends a message to the AI model and returns the response.
func (c *Client) SendMessage(ctx context.Context, text string) (string, error) {
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(text),
		}),
		Seed:  openai.Int(1),
		Model: openai.F(openai.ChatModelGPT4o),
	})

	if err != nil {
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}
