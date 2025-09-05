package analyzer

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestKeywordExtractor_ExtractKeywords(t *testing.T) {
	ke := NewKeywordExtractor()
	
	tests := []struct {
		name     string
		text     string
		topN     int
		expected int
	}{
		{
			name: "Extract keywords from technical text",
			text: `Machine learning is a subset of artificial intelligence that focuses on 
				   building systems that learn from data. Deep learning models use neural 
				   networks with multiple layers to process complex patterns in large datasets.
				   These models have revolutionized computer vision and natural language processing.`,
			topN:     3,
			expected: 3,
		},
		{
			name:     "Handle empty text",
			text:     "",
			topN:     3,
			expected: 0,
		},
		{
			name:     "Handle text with only stop words",
			text:     "the and or but with for to be",
			topN:     3,
			expected: 0,
		},
		{
			name: "Extract from business text",
			text: `The company announced a new product launch scheduled for next quarter. 
				   The marketing team has developed a comprehensive strategy to reach target 
				   customers through digital channels and social media platforms.`,
			topN:     3,
			expected: 3,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := ke.ExtractKeywords(tt.text, tt.topN)
			assert.LessOrEqual(t, len(keywords), tt.expected)
			
			if tt.text != "" && tt.expected > 0 {
				assert.NotEmpty(t, keywords)
				for _, keyword := range keywords {
					assert.NotEmpty(t, keyword)
					assert.Greater(t, len(keyword), 2)
				}
			}
		})
	}
}

func TestKeywordExtractor_IsLikelyNoun(t *testing.T) {
	ke := NewKeywordExtractor()
	
	tests := []struct {
		word     string
		expected bool
	}{
		{"organization", true},
		{"development", true},
		{"happiness", true},
		{"teacher", true},
		{"manager", true},
		{"democracy", true},
		{"friendship", true},
		{"technology", true},
		{"running", true},
		{"data", true},
		{"api", true},
		{"run", false},
		{"is", false},
		{"at", false},
		{"go", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			result := ke.isLikelyNoun(tt.word)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		summary   string
		topics    []string
		minScore  float64
		maxScore  float64
	}{
		{
			name: "High confidence - good text",
			text: `This is a comprehensive article about machine learning and artificial intelligence.
				   It covers various topics including neural networks, deep learning, and computer vision.
				   The applications range from healthcare to autonomous vehicles. Machine learning has
				   transformed how we process and analyze large datasets.`,
			summary: "This article discusses machine learning and AI applications.",
			topics:  []string{"machine learning", "artificial intelligence", "applications"},
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			name:      "Low confidence - empty inputs",
			text:      "",
			summary:   "",
			topics:    []string{},
			minScore:  0.0,
			maxScore:  0.0,
		},
		{
			name:      "Medium confidence - short text",
			text:      "This is a short text about testing.",
			summary:   "A text about testing.",
			topics:    []string{"testing"},
			minScore:  0.4,
			maxScore:  0.7,
		},
		{
			name: "Good confidence - normal text",
			text: `Software development involves writing code, testing, and deployment.
				   Modern practices include agile methodologies and continuous integration.`,
			summary: "Overview of software development practices.",
			topics:  []string{"development", "testing", "deployment"},
			minScore: 0.6,
			maxScore: 0.9,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := CalculateConfidence(tt.text, tt.summary, tt.topics)
			assert.GreaterOrEqual(t, confidence, tt.minScore)
			assert.LessOrEqual(t, confidence, tt.maxScore)
		})
	}
}

func BenchmarkKeywordExtraction(b *testing.B) {
	ke := NewKeywordExtractor()
	text := `Machine learning is a subset of artificial intelligence that focuses on 
			 building systems that learn from data. Deep learning models use neural 
			 networks with multiple layers to process complex patterns in large datasets.
			 These models have revolutionized computer vision and natural language processing.
			 Applications include image recognition, speech recognition, and language translation.`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ke.ExtractKeywords(text, 3)
	}
}