package irc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
)

type Client struct {
	conn *ircevent.Connection
}

type Option func(*Client)

func New(server string, channel string, nick string, options ...Option) *Client {
	client := &Client{
		conn: &ircevent.Connection{
			Server:      server,
			UseTLS:      false,
			Nick:        nick,
			Debug:       true,
			RequestCaps: []string{"server-time", "message-tags"},
		},
	}
	// Join default channel on connect
	client.conn.AddConnectCallback(func(e ircmsg.Message) {
		client.conn.Join(channel)
	})
	for _, opt := range options {
		opt(client)
	}
	return client
}

type GoogleSearch interface {
	ImageSearch(ctx context.Context, query string) (string, error)
	Search(ctx context.Context, query string) (string, error)
}

func WithGoogleSearch(search GoogleSearch) Option {
	ctx := context.Background()
	return func(c *Client) {
		// Handle google search with !g
		c.conn.AddCallback("PRIVMSG", func(e ircmsg.Message) {
			if trimmed, found := strings.CutPrefix(e.Params[1], "!g"); found {
				result, err := search.Search(ctx, trimmed)
				if err != nil {
					fmt.Fprintf(os.Stderr, "search error: %v", err)
					c.conn.Privmsg("toraton", fmt.Sprintf("search error: %v", err))
					return
				}
				if err := c.conn.Privmsg(e.Params[0], result); err != nil {
					fmt.Fprintf(os.Stderr, "irc error: %s: %v", result, err)
					return
				}
			} else if trimmed, found := strings.CutPrefix(e.Params[1], "!img"); found {
				result, err := search.ImageSearch(ctx, trimmed)
				if err != nil {
					fmt.Fprintf(os.Stderr, "search error: %v", err)
					c.conn.Privmsg("toraton", fmt.Sprintf("search error: %v", err))
					return
				}
				if err := c.conn.Privmsg(e.Params[0], result); err != nil {
					fmt.Fprintf(os.Stderr, "irc error: %s: %v", result, err)
					return
				}
			}
		})
	}
}

type GenAI interface {
	Chat(ctx context.Context, user, message string) (string, error)
}

func WithGenAI(genAI GenAI) Option {
	chat := func(c *Client, channel, user, msg string) {
		result, err := genAI.Chat(context.Background(), user, msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "genai error: %v", err)
			c.conn.Privmsg("toraton", fmt.Sprintf("genai error: %v", err))
			return
		}
		if err := c.conn.Privmsg(channel, result); err != nil {
			fmt.Fprintf(os.Stderr, "irc error: %s: %v", result, err)
			return
		}
	}
	return func(c *Client) {
		c.conn.AddCallback("PRIVMSG", func(e ircmsg.Message) {
			if trimmed, found := strings.CutPrefix(e.Params[1], "!ai"); found {
				chat(c, e.Params[0], e.Nick(), trimmed)
			} else if trimmed, found := strings.CutPrefix(e.Params[1], "!newchat"); found {
				chat(c, e.Params[0], e.Nick(), trimmed)
			} else if trimmed, found := strings.CutPrefix(e.Params[1], "!chat"); found {
				chat(c, e.Params[0], e.Nick(), trimmed)
			}
		})
	}
}

func (c *Client) Start() error {
	err := c.conn.Connect()
	if err != nil {
		return err
	}
	c.conn.Loop()
	return nil
}
