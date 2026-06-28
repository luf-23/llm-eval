package provider

import (
	"context"
	"fmt"
)

type Request struct {
	SuitePrompt string
	CaseID      string
	Input       string
}

type Response struct {
	Output string `json:"output"`
	Raw    string `json:"raw"`
}

type Provider interface {
	Name() string
	Generate(ctx context.Context, req Request) (Response, error)
}

func New(name string) (Provider, error) {
	switch name {
	case "", "deepseek":
		return NewDeepSeekFromEnv(), nil
	case "qwen":
		return NewQwenFromEnv(), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}
