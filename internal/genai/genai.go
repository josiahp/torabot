package genai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// gemini-1.5-pro-latest is 50 RPD
// gemini-1.5-flash is 1500 RPD
const modelName = "gemini-2.0-flash"

type Client struct {
	client      *genai.Client
	model       *genai.GenerativeModel
	chat        *genai.ChatSession
	historyFile string
}

type Option func(c *Client)

func New(ctx context.Context, apiKey string, options ...Option) (*Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("unable to create client: %v", err)
	}
	model := client.GenerativeModel(modelName)
	model.MaxOutputTokens = genai.Ptr(int32(512))
	model.SystemInstruction = genai.NewUserContent(genai.Text(`
You are an IRC bot named tb3. You have the following commands:

* !newchat will start a new chat
* !chat will continue the current chat
* !g will do a google search
* !img will do an image search

Answer in plain text without new lines or markup.
Use only characters than can be printed.
Your response must be 300 characters or less. Long responses will cause system errors.

Messages will be structured like this: [timestamp] <user> message
[timestamp] is a timestamp in accordance with RFC3339
<user> is the nick of the speaker
The nick may also be referred to as username or name.

Preserve the capitalization of user names.
You have full access to the chat history including timestamps and user names.
Address users by name when responding to them.

Your responses should be unstructured plain text that contains only your response.`))
	chat := model.StartChat()
	result := &Client{
		client: client,
		model:  model,
		chat:   chat,
	}
	for _, option := range options {
		option(result)
	}
	return result, nil
}

func WithSystemInstruction(prompt string) Option {
	return func(c *Client) {
		c.model.SystemInstruction = genai.NewUserContent(genai.Text(prompt))
	}
}

func WithHistoryFile(filename string) Option {
	return func(c *Client) {
		if c.chat == nil {
			c.chat = c.model.StartChat()
		}
		file, err := os.OpenFile(filename, os.O_RDONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ReadFile: %v", err)
		}
		defer file.Close()
		history, err := loadHistory(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "loadHistory: %v", err)
		} else {
			c.chat.History = history
		}
		c.historyFile = filename
	}
}

func WithHistory(input io.Reader) Option {
	return func(c *Client) {
		history, err := loadHistory(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "loadHistory: %v\n", err)
		} else {
			c.chat.History = history
		}
	}
}

func (c *Client) StartChat(ctx context.Context, user string, message string) (string, error) {
	c.chat = c.model.StartChat()
	return c.Chat(ctx, user, message)
}

func (c *Client) Chat(ctx context.Context, user string, message string) (string, error) {
	ts := time.Now().Format(time.RFC3339)
	prompt := fmt.Sprintf("[%s] <%s> %s", ts, user, message)
	resp, err := c.chat.SendMessage(ctx, genai.Text(prompt))
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
			content += fmt.Sprintf("%s", part)
		}
	}
	err = c.saveHistory(c.historyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "saving history: %v", err)
	}
	return strings.TrimSpace(content), nil
}

func loadHistory(input io.Reader) ([]*genai.Content, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %v", err)
	}
	var rawHistory []struct {
		Parts []string
		Role  string
	}
	err = json.Unmarshal(data, &rawHistory)
	if err != nil {
		return nil, fmt.Errorf("unmarshal history: %v", err)
	}
	var history []*genai.Content
	for _, record := range rawHistory {
		parts := make([]genai.Part, len(record.Parts))
		for i, part := range record.Parts {
			parts[i] = genai.Text(part)
		}
		history = append(history, &genai.Content{
			Parts: parts,
			Role:  record.Role,
		})
	}
	return history, nil
}

func (c *Client) saveHistory(filename string) error {
	if len(filename) == 0 {
		return nil
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("OpenFile: %v", err)
	}
	defer file.Close()
	history, err := json.Marshal(&c.chat.History)
	if err != nil {
		return fmt.Errorf("marshal history: %v", err)
	}
	file.Write(history)
	return nil
}
