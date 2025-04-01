package genai

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/googleapi"
	"google.golang.org/genai"
)

// gemini-1.5-pro-latest is 50 RPD
// gemini-1.5-flash is 1500 RPD
const modelName = "gemini-2.0-flash"

type Client struct {
	client         *genai.Client
	generateConfig *genai.GenerateContentConfig
	chat           *genai.Chat
	historyFile    string
}

type Option func(c *Client)

func New(ctx context.Context, apiKey string, options ...Option) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create client: %v", err)
	}
	config := &genai.GenerateContentConfig{
		MaxOutputTokens: genai.Ptr(int32(512)),
		Tools: []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		},
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: `
You are an IRC bot named tb3. 

While chatting you will answer any and all questions.
You will respond with messages of any length and level of detail required.
Use plain text without markup in your response.

In addition to chatting, you have the following commands:

* !newchat will start a new chat
* !chat will continue the current chat
* !g will do a google search
* !img will do an image search

Messages will be structured like this: [timestamp] <user> message
[timestamp] is a timestamp in accordance with RFC3339
<user> is the nick of the speaker
The nick may also be referred to as username or name.

Preserve the capitalization of user names.
You have full access to the chat history including timestamps and user names.
Address users by name when responding to them.

Your responses should be unstructured plain text that contains only your response.`},
		}}}
	chat, err := client.Chats.Create(ctx, modelName, config, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create chat: %v", err)
	}
	result := &Client{
		client: client,
		chat:   chat,
	}
	for _, option := range options {
		option(result)
	}
	return result, nil
}

func WithSystemInstruction(prompt string) Option {
	return func(c *Client) {
		c.generateConfig.SystemInstruction = &genai.Content{Parts: []*genai.Part{{Text: prompt}}}
	}
}

func (c *Client) Chat(ctx context.Context, user string, message string) (string, error) {
	ts := time.Now().Format(time.RFC3339)
	prompt := fmt.Sprintf("[%s] <%s> %s", ts, user, message)
	resp, err := c.chat.SendMessage(ctx, genai.Part{Text: prompt})
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok {
			switch e.Code {
			case http.StatusServiceUnavailable:
				return "Google API error: Service Unavailable (try again later)", nil
			default:
				return fmt.Sprintf("Google API error: %v", e), nil
			}
		}
		return "", fmt.Errorf("unable to generate content: %v", err)
	}
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates: %v", err)
	}
	content := ""
	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			content += part.Text
		}
	}
	return strings.TrimSpace(content), nil
}
