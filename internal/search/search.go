package search

import (
	"context"

	"google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

type Client struct {
	service *customsearch.Service

	// APIKey is the API Key of the Custom Search Engine
	APIKey string
	// ID is the ID of the Custom Search Engine
	ID string
}

func New(ctx context.Context, apiKey string, id string) (*Client, error) {
	service, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &Client{
		service: service,
		APIKey:  apiKey,
		ID:      id,
	}, nil
}

type searchType string

const searchType_undefined searchType = "searchTypeUndefined"
const searchType_image searchType = "image"

func (c *Client) search(ctx context.Context, q string, t searchType) (string, error) {
	search, err := c.service.Cse.List().Context(ctx).Cx(c.ID).Q(q).SearchType(string(t)).Do()
	if err != nil {
		return "", err
	}
	if len(search.Items) > 0 {
		return search.Items[0].Link, nil
	}
	return "", nil
}

func (c *Client) ImageSearch(ctx context.Context, q string) (string, error) {
	return c.search(ctx, q, searchType_image)
}

func (c *Client) Search(ctx context.Context, q string) (string, error) {
	return c.search(ctx, q, searchType_undefined)
}
