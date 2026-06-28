package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ChatProvider struct {
	name    string
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

func NewDeepSeekFromEnv() ChatProvider {
	return ChatProvider{
		name:    "deepseek",
		apiKey:  os.Getenv("DEEPSEEK_API_KEY"),
		model:   envOrDefault("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		baseURL: strings.TrimRight(envOrDefault("DEEPSEEK_BASE_URL", "https://api.deepseek.com"), "/"),
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

func NewQwenFromEnv() ChatProvider {
	apiKey := os.Getenv("QWEN_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("DASHSCOPE_API_KEY")
	}
	return ChatProvider{
		name:    "qwen",
		apiKey:  apiKey,
		model:   envOrDefault("QWEN_MODEL", "qwen-plus"),
		baseURL: strings.TrimRight(envOrDefault("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"), "/"),
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

func (p ChatProvider) Name() string { return p.name }

func (p ChatProvider) Generate(ctx context.Context, req Request) (Response, error) {
	if p.apiKey == "" {
		return Response{}, fmt.Errorf("%s api key is required", strings.ToUpper(p.name))
	}

	payload := chatRequest{
		Model: p.model,
		Messages: []chatMessage{
			{Role: "system", Content: req.SuitePrompt},
			{Role: "user", Content: req.Input},
		},
		Stream: false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("%s status %d: %s", p.name, resp.StatusCode, string(raw))
	}

	output := extractChatOutput(raw)
	if output == "" {
		return Response{}, fmt.Errorf("%s response does not contain choices[0].message.content", p.name)
	}
	return Response{Output: output, Raw: string(raw)}, nil
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func extractChatOutput(raw []byte) string {
	var envelope struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return ""
	}
	if len(envelope.Choices) == 0 {
		return ""
	}
	return envelope.Choices[0].Message.Content
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
