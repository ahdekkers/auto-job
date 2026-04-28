package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OllamaClient is a small wrapper around the Ollama generate API.
type OllamaClient struct {
	BaseURL string
	Client  *http.Client
}

// NewOllamaClient creates a new Ollama client.
// Example BaseURL: http://localhost:11434
func NewOllamaClient(baseURL string) *OllamaClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://192.168.1.97:11434"
	}

	return &OllamaClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type ollamaGenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// Generate sends a prompt to Ollama and returns the model response text.
func (c *OllamaClient) Generate(ctx context.Context, model, prompt string) (string, error) {
	if strings.TrimSpace(model) == "" {
		return "", fmt.Errorf("model is required")
	}
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt is required")
	}

	reqBody := ollamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request body: %w", err)
	}

	endpoint := c.BaseURL + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ollamaGenerateResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return "", fmt.Errorf("ollama api error: %s", errResp.Error)
		}
		return "", fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	var out ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if strings.TrimSpace(out.Error) != "" {
		return "", fmt.Errorf("ollama api error: %s", out.Error)
	}

	return out.Response, nil
}
