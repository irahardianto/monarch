package gemini

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"github.com/monarch-dev/monarch/internal/llm"
)

type Client struct {
	genaiClient *genai.Client
	modelName   string
}

func NewClient(apiKey string) (llm.Client, error) {
	ctx := context.Background()
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &Client{genaiClient: c, modelName: "gemini-pro"}, nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	model := c.genaiClient.GenerativeModel(c.modelName)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}
	
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}
	
	var output string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			output += string(txt)
		}
	}
	return output, nil
}

func (c *Client) Close() error {
	return c.genaiClient.Close()
}
