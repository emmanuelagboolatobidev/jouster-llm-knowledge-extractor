package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrEmptyInput     = errors.New("empty input text")
	ErrLLMUnavailable = errors.New("LLM service unavailable")
	ErrInvalidJSON    = errors.New("invalid JSON response from LLM")
)

type Provider interface {
	Analyze(ctx context.Context, text string) (*AnalysisResult, error)
	IsAvailable() bool
}

type AnalysisResult struct {
	Summary   string   `json:"summary"`
	Title     string   `json:"title"`
	Topics    []string `json:"topics"`
	Sentiment string   `json:"sentiment"`
}

type Config struct {
	Provider       string
	Model          string
	MaxTokens      int
	Temperature    float32
}

func NewProvider(config Config) (Provider, error) {
	switch config.Provider {
	case "mock":
		return NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}
}

func parseJSONResponse(content string) (*AnalysisResult, error) {
	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	
	if result.Summary == "" {
		result.Summary = "No summary available"
	}
	
	if len(result.Topics) == 0 {
		result.Topics = []string{"general", "uncategorized", "text"}
	} else if len(result.Topics) > 3 {
		result.Topics = result.Topics[:3]
	}
	
	validSentiments := map[string]bool{
		"positive": true,
		"neutral":  true,
		"negative": true,
	}
	
	if !validSentiments[result.Sentiment] {
		result.Sentiment = "neutral"
	}
	
	return &result, nil
}