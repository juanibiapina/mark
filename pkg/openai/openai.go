package openai

import (
	"mark/pkg/model"

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

// CompleteStreaming sends a list of messages to the OpenAI API and streams the response
func (a *OpenAI) CompleteStreaming(c *model.Thread, s *model.StreamingMessage, p model.Project) error {
	ctx := s.Ctx
	pch := s.Chunks
	ch := s.Reply

	defer close(pch)
	defer close(ch)

	// Initialize the chat messages
	var chatMessages []openai.ChatCompletionMessageParamUnion

	// Add project prompt
	tmp, err := p.Prompt()
	if err != nil {
		return err
	}
	if tmp != "" {
		chatMessages = append(chatMessages, openai.UserMessage(tmp))
	}

	// Add Pull Request prompt
	tmp = c.PullRequest.Prompt()
	if tmp != "" {
		chatMessages = append(chatMessages, openai.UserMessage(tmp))
	}

	// Add the messages
	for _, msg := range c.Messages {
		if msg.Role == model.RoleUser {
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

		if refusal, ok := acc.JustFinishedRefusal(); ok {
			println("Refusal stream finished:", refusal)
		}

		// it's best to use chunks after handling JustFinished events
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				pch <- chunk.Choices[0].Delta.Content
			}
		}
	}

	if err := stream.Err(); err != nil {
		return err
	}

	response := acc.Choices[0].Message.Content
	ch <- response
	return nil
}
