package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"example.com/chat_app/chat_service/structs"
)

// AiAssistantClient is an http client wrapper for communication with media service.
type AiAssistantClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewAiClient creates a new AiAssistantClient.
// It reads the base URL for the media service from the AI_ASSISTANT_URL environment variable.
func NewAiClient() (*AiAssistantClient, error) {
	baseURL := os.Getenv("AI_ASSISTANT_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("AI_ASSISTANT_URL environment variable not set")
	}

	return &AiAssistantClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}, nil
}

// GetMessagesSummary sends a POST request to the AI assistant to get a summary of the messages.
func (c *AiAssistantClient) GetMessagesSummary(ctx context.Context, messages []structs.MessageDto) (*structs.MessagesSummary, error) {
	payload := map[string][]structs.MessageDto{
		"messages": messages,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal messages: %w", err)
	}

	fmt.Printf("Request: %s\n", string(jsonData))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/summarize_chat", c.BaseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Response: %s\n", string(bodyBytes))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI assistant returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var summary structs.MessagesSummary
	if err := json.Unmarshal(bodyBytes, &summary); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &summary, nil
}
