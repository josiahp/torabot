# torabot (tb3)

torabot is an IRC bot maintained by #codelove on SlashNET. torabot was created as a proof of concept for a minimal IRC bot in go. Use at your own risk.

## Features

torabot has the following features:

* Google Search with `!g`
* Google Image Search with `!img`
* AI chat with `!chat`

## Configuration

torabot is configured via environment variables. I use [direnv](https://direnv.net/) to set them.

* SEARCH_API_KEY: Google Custom Search API Key
* SEARCH_ID: Google Custom Search Engine ID
* GEMINI_API_KEY: Google Gemini API Key

## Building (binary)

```console
go build .
```

## Building (docker)

```console
docker build . -t torabot:$(git describe --exact-match --tags)
```

## Running (binary)

```console
./torabot
```

## Running (docker)

```console
docker run -d \
  -e SEARCH_API_KEY \
  -e SEARCH_ID \
  -e GEMINI_API_KEY \
  --name torabot \
  --mount src=torabot,dst=/data \
  torabot:latest
```