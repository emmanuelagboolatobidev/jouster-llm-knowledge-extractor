package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/user/llm-knowledge-extractor/internal/database"
	"github.com/user/llm-knowledge-extractor/internal/handlers"
	"github.com/user/llm-knowledge-extractor/internal/llm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/knowledge.db"
	}
	
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}
	
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
	llmConfig := llm.Config{
		Provider:     os.Getenv("LLM_PROVIDER"),
	}
	
	if llmConfig.Provider == "" {
		llmConfig.Provider = "mock"
		log.Println("No LLM provider configured, using mock provider")
	}
	
	llmProvider, err := llm.NewProvider(llmConfig)
	if err != nil {
		log.Fatalf("Failed to initialize LLM provider: %v", err)
	}
	
	handler := handlers.New(db, llmProvider)
	
	r := gin.Default()
	
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	
	r.POST("/analyze", handler.AnalyzeText)
	r.POST("/batch-analyze", handler.BatchAnalyzeText)
	r.GET("/search", handler.SearchAnalyses)
	
	log.Printf("Starting server on port %s", port)
	log.Printf("Database path: %s", dbPath)
	log.Printf("LLM Provider: %s", llmConfig.Provider)
	
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}