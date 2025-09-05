package llm

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestMockProvider_Analyze(t *testing.T) {
	provider := NewMockProvider()
	
	tests := []struct {
		name        string
		text        string
		expectError bool
	}{
		{
			name: "Analyze normal text",
			text: `Artificial intelligence is transforming industries across the globe. 
				   From healthcare to finance, AI applications are improving efficiency 
				   and enabling new capabilities.`,
			expectError: false,
		},
		{
			name:        "Handle empty text",
			text:        "",
			expectError: true,
		},
		{
			name:        "Handle whitespace only",
			text:        "   \n\t  ",
			expectError: true,
		},
		{
			name:        "Analyze short text",
			text:        "This is a test.",
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := provider.Analyze(ctx, tt.text)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				if err == nil {
					assert.NotNil(t, result)
					assert.NotEmpty(t, result.Summary)
					assert.Len(t, result.Topics, 3)
					assert.Contains(t, []string{"positive", "neutral", "negative"}, result.Sentiment)
				}
			}
		})
	}
}


func TestParseJSONResponse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		validate    func(*testing.T, *AnalysisResult)
	}{
		{
			name: "Valid complete JSON",
			input: `{
				"summary": "This is a summary",
				"title": "Test Title",
				"topics": ["topic1", "topic2", "topic3"],
				"sentiment": "positive"
			}`,
			expectError: false,
			validate: func(t *testing.T, result *AnalysisResult) {
				assert.Equal(t, "This is a summary", result.Summary)
				assert.Equal(t, "Test Title", result.Title)
				assert.Equal(t, []string{"topic1", "topic2", "topic3"}, result.Topics)
				assert.Equal(t, "positive", result.Sentiment)
			},
		},
		{
			name: "Missing summary defaults",
			input: `{
				"title": "Test",
				"topics": ["topic1"],
				"sentiment": "neutral"
			}`,
			expectError: false,
			validate: func(t *testing.T, result *AnalysisResult) {
				assert.Equal(t, "No summary available", result.Summary)
			},
		},
		{
			name: "Too many topics truncated",
			input: `{
				"summary": "Summary",
				"topics": ["t1", "t2", "t3", "t4", "t5"],
				"sentiment": "negative"
			}`,
			expectError: false,
			validate: func(t *testing.T, result *AnalysisResult) {
				assert.Len(t, result.Topics, 3)
				assert.Equal(t, []string{"t1", "t2", "t3"}, result.Topics)
			},
		},
		{
			name: "Invalid sentiment defaults to neutral",
			input: `{
				"summary": "Summary",
				"topics": ["t1"],
				"sentiment": "invalid"
			}`,
			expectError: false,
			validate: func(t *testing.T, result *AnalysisResult) {
				assert.Equal(t, "neutral", result.Sentiment)
			},
		},
		{
			name: "Empty topics gets defaults",
			input: `{
				"summary": "Summary",
				"topics": [],
				"sentiment": "positive"
			}`,
			expectError: false,
			validate: func(t *testing.T, result *AnalysisResult) {
				assert.Equal(t, []string{"general", "uncategorized", "text"}, result.Topics)
			},
		},
		{
			name:        "Invalid JSON",
			input:       `{invalid json}`,
			expectError: true,
		},
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONResponse(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		expectType  string
	}{
		{
			name: "Create Mock provider",
			config: Config{
				Provider: "mock",
			},
			expectError: false,
			expectType:  "*llm.MockProvider",
		},
		{
			name: "Unknown provider fails",
			config: Config{
				Provider: "unknown",
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}