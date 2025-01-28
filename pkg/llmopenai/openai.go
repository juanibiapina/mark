package llmopenai

import (
	"context"
	"encoding/json"

	"ant/pkg/llm"

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
func (a *OpenAI) CompleteStreaming(c *llm.Conversation, s *llm.StreamingMessage) error {
	ctx := s.Ctx
	pch := s.Chunks
	ch := s.Reply

	defer close(pch)
	defer close(ch)

	// Initialize the chat messages
	var chatMessages []openai.ChatCompletionMessageParamUnion

	// Add a user message containing context
	context := c.Context()
	if len(context) != 0 {
		con := "Context:\n"
		for _, v := range context {
			con += v + "\n"
		}
		chatMessages = append(chatMessages, openai.UserMessage(con))
	}

	// Add the actual conversation messages
	for _, msg := range c.Messages() {
		if msg.Role == llm.RoleUser {
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

		// if using tool calls
		if _, ok := acc.JustFinishedToolCall(); ok {
		}

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

func (a *OpenAI) CompleteStructured(c *llm.Conversation, rs llm.ResponseSchema, v interface{}) error {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F(rs.Name),
		Description: openai.F(rs.Description),
		Schema:      openai.F(rs.Schema),
		Strict:      openai.Bool(true),
	}

	// Initialize the chat messages
	var chatMessages []openai.ChatCompletionMessageParamUnion

	// Add a user message containing convContext
	convContext := c.Context()
	if len(convContext) != 0 {
		con := "Context:\n"
		for _, v := range convContext {
			con += v + "\n"
		}
		chatMessages = append(chatMessages, openai.UserMessage(con))
	}

	// Add the actual conversation messages
	for _, msg := range c.Messages() {
		if msg.Role == llm.RoleUser {
			chatMessages = append(chatMessages, openai.UserMessage(msg.Content))
		} else {
			chatMessages = append(chatMessages, openai.AssistantMessage(msg.Content))
		}
	}

	chat, err := a.client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: openai.F(chatMessages),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		panic(err.Error())
	}

	// The model responds with a JSON string, so parse it into a struct
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), v)
	if err != nil {
		panic(err.Error())
	}

	return nil
}
