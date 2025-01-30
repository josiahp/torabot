package main

import (
	"context"
	"fmt"
	"os"

	"codelove.org/torabot/internal/genai"
	"codelove.org/torabot/internal/irc"
	"codelove.org/torabot/internal/search"
)

const (
	defaultNick = "tb3"
)

func main() {
	nick := os.Getenv("IRC_NICK")
	if len(nick) == 0 {
		nick = defaultNick
	}
	apiKey := os.Getenv("SEARCH_API_KEY")
	if len(apiKey) == 0 {
		fmt.Fprintf(os.Stderr, "SEARCH_API_KEY must be set to a Google Custom Search Engine API Key")
		os.Exit(1)
	}
	id := os.Getenv("SEARCH_ID")
	if len(id) == 0 {
		fmt.Fprintf(os.Stderr, "SEARCH_ID must be set to a Google Custom Search Engine ID")
		os.Exit(1)
	}
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if len(id) == 0 {
		fmt.Fprintf(os.Stderr, "GEMINI_API_KEY must be set to a Google Gemini API Key")
		os.Exit(1)
	}
	ctx := context.Background()
	search, err := search.New(ctx, apiKey, id)
	if err != nil {
		panic(err)
	}
	genai, err := genai.New(ctx, geminiKey)
	if err != nil {
		panic(err)
	}
	irc := irc.New("irc.slashnet.org:6667", "#codelove", nick,
		irc.WithGoogleSearch(search),
		irc.WithGenAI(genai))
	irc.Start()
}
