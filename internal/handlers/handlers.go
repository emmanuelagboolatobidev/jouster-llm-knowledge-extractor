package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/user/llm-knowledge-extractor/internal/analyzer"
	"github.com/user/llm-knowledge-extractor/internal/database"
	"github.com/user/llm-knowledge-extractor/internal/llm"
	"github.com/user/llm-knowledge-extractor/internal/models"
)

type Handler struct {
	db               *database.DB
	llmProvider      llm.Provider
	keywordExtractor *analyzer.KeywordExtractor
}

func New(db *database.DB, llmProvider llm.Provider) *Handler {
	return &Handler{
		db:               db,
		llmProvider:      llmProvider,
		keywordExtractor: analyzer.NewKeywordExtractor(),
	}
}

func (h *Handler) AnalyzeText(c *gin.Context) {
	var req models.AnalyzeRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}
	
	if req.Text == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Text cannot be empty",
			Code:  "EMPTY_INPUT",
		})
		return
	}
	
	startTime := time.Now()
	
	ctx, cancel := context.WithTimeout(c.Request.Context(), 45*time.Second)
	defer cancel()
	
	llmResult, err := h.llmProvider.Analyze(ctx, req.Text)
	if err != nil {
		if err == llm.ErrEmptyInput {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Text cannot be empty",
				Code:  "EMPTY_INPUT",
			})
			return
		}
		
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error:   "LLM service unavailable",
			Code:    "LLM_UNAVAILABLE",
			Details: err.Error(),
		})
		return
	}
	
	keywords := h.keywordExtractor.ExtractKeywords(req.Text, 3)
	
	metadata := map[string]interface{}{
		"title":     llmResult.Title,
		"topics":    llmResult.Topics,
		"sentiment": llmResult.Sentiment,
		"keywords":  keywords,
	}
	
	confidence := analyzer.CalculateConfidence(req.Text, llmResult.Summary, llmResult.Topics)
	
	analysis := &models.TextAnalysis{
		ID:           uuid.New().String(),
		Text:         req.Text,
		Summary:      llmResult.Summary,
		Metadata:     metadata,
		Confidence:   confidence,
		CreatedAt:    time.Now(),
		ProcessingMS: time.Since(startTime).Milliseconds(),
	}
	
	if err := h.db.SaveAnalysis(analysis); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to save analysis",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.AnalyzeResponse{
		ID:         analysis.ID,
		Summary:    analysis.Summary,
		Metadata:   analysis.Metadata,
		Confidence: analysis.Confidence,
	})
}

func (h *Handler) BatchAnalyzeText(c *gin.Context) {
	var req models.BatchAnalyzeRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}
	
	if len(req.Texts) == 0 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "No texts provided",
			Code:  "EMPTY_INPUT",
		})
		return
	}
	
	if len(req.Texts) > 10 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Maximum 10 texts allowed per batch",
			Code:  "BATCH_SIZE_EXCEEDED",
		})
		return
	}
	
	var wg sync.WaitGroup
	results := make([]models.AnalyzeResponse, len(req.Texts))
	errors := make([]models.BatchError, 0)
	var errorsMu sync.Mutex
	
	semaphore := make(chan struct{}, 3)
	
	for i, text := range req.Texts {
		wg.Add(1)
		go func(index int, textContent string) {
			defer wg.Done()
			
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			if textContent == "" {
				errorsMu.Lock()
				errors = append(errors, models.BatchError{
					Index: index,
					Error: "Text cannot be empty",
				})
				errorsMu.Unlock()
				return
			}
			
			startTime := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			
			llmResult, err := h.llmProvider.Analyze(ctx, textContent)
			if err != nil {
				errorsMu.Lock()
				errors = append(errors, models.BatchError{
					Index: index,
					Error: fmt.Sprintf("Analysis failed: %v", err),
				})
				errorsMu.Unlock()
				return
			}
			
			keywords := h.keywordExtractor.ExtractKeywords(textContent, 3)
			
			metadata := map[string]interface{}{
				"title":     llmResult.Title,
				"topics":    llmResult.Topics,
				"sentiment": llmResult.Sentiment,
				"keywords":  keywords,
			}
			
			confidence := analyzer.CalculateConfidence(textContent, llmResult.Summary, llmResult.Topics)
			
			analysis := &models.TextAnalysis{
				ID:           uuid.New().String(),
				Text:         textContent,
				Summary:      llmResult.Summary,
				Metadata:     metadata,
				Confidence:   confidence,
				CreatedAt:    time.Now(),
				ProcessingMS: time.Since(startTime).Milliseconds(),
			}
			
			if err := h.db.SaveAnalysis(analysis); err != nil {
				errorsMu.Lock()
				errors = append(errors, models.BatchError{
					Index: index,
					Error: fmt.Sprintf("Failed to save: %v", err),
				})
				errorsMu.Unlock()
				return
			}
			
			results[index] = models.AnalyzeResponse{
				ID:         analysis.ID,
				Summary:    analysis.Summary,
				Metadata:   analysis.Metadata,
				Confidence: analysis.Confidence,
			}
		}(i, text)
	}
	
	wg.Wait()
	
	successResults := make([]models.AnalyzeResponse, 0)
	for _, result := range results {
		if result.ID != "" {
			successResults = append(successResults, result)
		}
	}
	
	c.JSON(http.StatusOK, models.BatchAnalyzeResponse{
		Results: successResults,
		Failed:  errors,
	})
}

func (h *Handler) SearchAnalyses(c *gin.Context) {
	var query models.SearchQuery
	
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid query parameters",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}
	
	if query.Limit == 0 {
		query.Limit = 50
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	
	analyses, err := h.db.SearchAnalyses(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Search failed",
			Code:    "DB_ERROR",
			Details: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"results": analyses,
		"count":   len(analyses),
		"query":   query,
	})
}