package ai

import (
	"ant/pkg/model"
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

// Complete sends a list of messages to the OpenAI API and returns the response
func (c *Client) Complete(ctx context.Context, messages []model.Message, pch chan string, ch chan string) {
	// Convert messages to OpenAI format
	var chatMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range messages {
		if msg.Role == model.User {
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		} else {
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		}
	}

	stream := c.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(chatMessages),
		Seed:     openai.Int(1),
		Model:    openai.F(openai.ChatModelGPT4o),
	})

	acc := openai.ChatCompletionAccumulator{}

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		if _, ok := acc.JustFinishedContent(); ok {
			// unused for now
		}

		// if using tool calls
		if tool, ok := acc.JustFinishedToolCall(); ok {
			println("Tool call stream finished:", tool.Index, tool.Name, tool.Arguments)
		}

		if refusal, ok := acc.JustFinishedRefusal(); ok {
			println("Refusal stream finished:", refusal)
		}

		// it's best to use chunks after handling JustFinished events
		if len(chunk.Choices) > 0 {
			pch <- chunk.Choices[0].Delta.Content
		}
	}

	if err := stream.Err(); err != nil {
		panic(err)
	}

	close(pch)

	// After the stream is finished, acc can be used like a ChatCompletion
	response := acc.Choices[0].Message.Content

	ch <- response
	close(ch)
}
