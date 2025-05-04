package providers

import (
	"context"
	"log/slog"

	"mark/internal/llm/provider"
	"mark/internal/model"

	"github.com/openai/openai-go"
)

type OpenAI struct {
	client openai.Client
}

func NewOpenAIClient() *OpenAI {
	return &OpenAI{
		client: openai.NewClient(),
	}
}

// CompleteStreaming sends a list of messages to the OpenAI API and streams the response
func (a *OpenAI) CompleteStreaming(ctx context.Context, c model.Thread) (<-chan provider.StreamingEvent, error) {
	slog.Info("Starting streaming completion")

	eventCh := make(chan provider.StreamingEvent)

	go func() {
		// Initialize the chat messages
		var chatMessages []openai.ChatCompletionMessageParamUnion

		// Add the messages
		for _, msg := range c.Messages {
			if msg.Role == model.RoleUser {
				chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
			} else {
				chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
			}
		}

		stream := a.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages: chatMessages,
			Seed:     openai.Int(1),
			Model:    openai.ChatModelGPT4o,
		})

		acc := openai.ChatCompletionAccumulator{}

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			if refusal, ok := acc.JustFinishedRefusal(); ok {
				println("Refusal stream finished:", refusal)
			}

			// it's best to use chunks after handling JustFinished events
			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					eventCh <- provider.StreamingEventChunk{Chunk: content}
				}
			}
		}

		if err := stream.Err(); err != nil {
			if err == context.Canceled {
				slog.Info("Streaming canceled")
				eventCh <- provider.StreamingCancelled{}
				return
			}
		}

		response := acc.Choices[0].Message.Content
		eventCh <- provider.StreamingEventEnd{Message: response}

		slog.Info("Streaming finished")
	}()

	return eventCh, nil
}
