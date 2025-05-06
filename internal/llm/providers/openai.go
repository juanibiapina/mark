package providers

import (
	"context"
	"log/slog"

	"mark/internal/llm"
	"mark/internal/llm/provider"
	"mark/internal/logging"

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

func convertMessages(messages []llm.Message) []openai.ChatCompletionMessageParamUnion {
	var chatMessages []openai.ChatCompletionMessageParamUnion

	for _, msg := range messages {
		if msg.Role == llm.RoleUser {
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		} else {
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		}
	}

	return chatMessages
}

func (a *OpenAI) CompleteStreaming(ctx context.Context, session llm.Session) (<-chan provider.StreamingEvent, error) {
	a.logger.Info("Starting streaming completion")

	eventCh := make(chan provider.StreamingEvent)

	go func() {
		messages := convertMessages(session.Messages)

		stream := a.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages: messages,
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
