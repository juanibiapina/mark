package ai

import (
	"context"

	"github.com/openai/openai-go"
)

type Message struct {
	Role    string
	Content string
}

type Client struct {
	client   *openai.Client
	messages []Message
}

func NewClient() *Client {
	return &Client{
		client:   openai.NewClient(),
		messages: make([]Message, 0),
	}
}

// SendMessage sends a message to the AI model and returns the response.
func (c *Client) SendMessage(ctx context.Context, text string) (string, error) {
	// Add user message to history
	c.messages = append(c.messages, Message{Role: "user", Content: text})

	// Convert messages to OpenAI format
	var chatMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range c.messages {
		if msg.Role == "user" {
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

	// Add assistant response to history
	response := completion.Choices[0].Message.Content
	c.messages = append(c.messages, Message{Role: "assistant", Content: response})

	return response, nil
}
