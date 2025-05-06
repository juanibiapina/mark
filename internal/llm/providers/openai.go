package providers

import (
	"context"
	"log/slog"

	"mark/internal/llm/provider"
	"mark/internal/logging"
	"mark/internal/model"

	"github.com/openai/openai-go"
)

type OpenAI struct {
	client openai.Client
	logger *slog.Logger
}

func NewOpenAIClient() *OpenAI {
	return &OpenAI{
		client: openai.NewClient(),
		logger: logging.NewLogger("provider-openai"),
	}
}

// CompleteStreaming sends a list of messages to the OpenAI API and streams the response
func (a *OpenAI) CompleteStreaming(ctx context.Context, session model.Session) (<-chan provider.StreamingEvent, error) {
	a.logger.Info("Starting streaming completion")

	eventCh := make(chan provider.StreamingEvent)

	go func() {
		// Initialize the chat messages
		var chatMessages []openai.ChatCompletionMessageParamUnion

		// Add the messages
		for _, msg := range session.Messages {
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
				a.logger.Info("Stream refusal", slog.String("refusal", refusal))
			}

			// it's best to use chunks after handling JustFinished events
			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					eventCh <- provider.StreamEventChunk{Chunk: content}
				}
			}
		}

		if err := stream.Err(); err != nil {
			if err == context.Canceled {
				a.logger.Info("Streaming canceled")
				return
			} else {
				a.logger.Error("Streaming error", slog.String("error", err.Error()))
				eventCh <- provider.StreamEventError{Error: err}
				return
			}
		}

		response := acc.Choices[0].Message.Content
		eventCh <- provider.StreamEventEnd{Message: response}

		a.logger.Info("Streaming finished")
	}()

	return eventCh, nil
}
