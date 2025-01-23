package llm

import (
	"context"

	"github.com/openai/openai-go"
)

type OpenAI struct {
	client *openai.Client
}

func NewOpenAIClient() *OpenAI {
	return &OpenAI{
		client: openai.NewClient(),
	}
}

// Complete sends a list of messages to the OpenAI API and returns the response
func (a *OpenAI) CompleteStreaming(ctx context.Context, c *Conversation, pch chan string, ch chan string) error {
	defer close(pch)
	defer close(ch)

	// Convert messages to OpenAI format
	var chatMessages []openai.ChatCompletionMessageParamUnion
	for _, msg := range c.Messages() {
		if msg.Role == RoleUser {
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		} else {
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		}
	}

	stream := a.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
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
		return err
	}

	// After the stream is finished, acc can be used like a ChatCompletion
	response := acc.Choices[0].Message.Content

	ch <- response
	return nil
}
