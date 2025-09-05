package models

import (
	"time"
)

type TextAnalysis struct {
	ID           string                 `json:"id" db:"id"`
	Text         string                 `json:"text" db:"text"`
	Summary      string                 `json:"summary" db:"summary"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	Confidence   float64                `json:"confidence" db:"confidence"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	ProcessingMS int64                  `json:"processing_ms" db:"processing_ms"`
}

type AnalysisMetadata struct {
	Title     string   `json:"title"`
	Topics    []string `json:"topics"`
	Sentiment string   `json:"sentiment"`
	Keywords  []string `json:"keywords"`
}

type AnalyzeRequest struct {
	Text string `json:"text" binding:"required,min=1"`
}

type BatchAnalyzeRequest struct {
	Texts []string `json:"texts" binding:"required,min=1,dive,min=1"`
}

type AnalyzeResponse struct {
	ID         string                 `json:"id"`
	Summary    string                 `json:"summary"`
	Metadata   map[string]interface{} `json:"metadata"`
	Confidence float64                `json:"confidence"`
}

type BatchAnalyzeResponse struct {
	Results []AnalyzeResponse `json:"results"`
	Failed  []BatchError      `json:"failed,omitempty"`
}

type BatchError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type SearchQuery struct {
	Topic   string `form:"topic"`
	Keyword string `form:"keyword"`
	Limit   int    `form:"limit,default=50"`
	Offset  int    `form:"offset,default=0"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}