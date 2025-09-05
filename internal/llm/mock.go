package llm

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type MockProvider struct {
	failureRate float64
	delay       time.Duration
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		failureRate: 0.1,
		delay:       100 * time.Millisecond,
	}
}

func (p *MockProvider) Analyze(ctx context.Context, text string) (*AnalysisResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, ErrEmptyInput
	}
	
	time.Sleep(p.delay)
	
	if rand.Float64() < p.failureRate {
		return nil, fmt.Errorf("%w: mock failure", ErrLLMUnavailable)
	}
	
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	
	words := strings.Fields(text)
	summaryLength := len(words) / 10
	if summaryLength < 5 {
		summaryLength = 5
	}
	if summaryLength > 20 {
		summaryLength = 20
	}
	
	summary := "This text discusses "
	if len(words) > 0 {
		summary += strings.Join(words[:min(summaryLength, len(words))], " ")
		summary += "..."
	}
	
	topics := []string{
		"technology",
		"innovation",
		"analysis",
	}
	
	sentiments := []string{"positive", "neutral", "negative"}
	sentiment := sentiments[rand.Intn(len(sentiments))]
	
	title := ""
	if len(words) > 3 {
		title = strings.Title(strings.Join(words[:min(3, len(words))], " "))
	}
	
	return &AnalysisResult{
		Summary:   summary,
		Title:     title,
		Topics:    topics,
		Sentiment: sentiment,
	}, nil
}

func (p *MockProvider) IsAvailable() bool {
	return rand.Float64() > 0.05
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}