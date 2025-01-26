package llm

import (
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
func (a *OpenAI) CompleteStreaming(c *Conversation, s *StreamingMessage) error {
	ctx := s.Ctx
	pch := s.Chunks
	ch := s.Reply

	defer close(pch)
	defer close(ch)

	// Initialize the chat messages
	var chatMessages []openai.ChatCompletionMessageParamUnion

	// Add a user message containing context
	if len(c.context) != 0 {
		con := "Context:\n"
		for _, v := range c.context {
			con += v + "\n"
		}
		chatMessages = append(chatMessages, openai.UserMessage(con))
	}

	// Add the actual conversation messages
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
		Tools: openai.F([]openai.ChatCompletionToolParam{
			{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name:        openai.String("ping"),
					Description: openai.String("A simple tool that returns 'pong'"),
					Strict:      openai.Bool(false),
				}),
			},
		}),
	})

	acc := openai.ChatCompletionAccumulator{}

	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		// if using tool calls
		if toolCall, ok := acc.JustFinishedToolCall(); ok {
			if toolCall.Name == "ping" {
				tool := &PingTool{}
				ch <- tool.Invoke()
				continue
			}
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

	response := acc.Choices[0].Message.Content
	ch <- response
	return nil
}
