package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	
	_ "github.com/mattn/go-sqlite3"
	"github.com/user/llm-knowledge-extractor/internal/models"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	db := &DB{conn: conn}
	
	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}
	
	return db, nil
}

func (db *DB) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS analyses (
		id TEXT PRIMARY KEY,
		text TEXT NOT NULL,
		summary TEXT NOT NULL,
		metadata TEXT NOT NULL,
		confidence REAL NOT NULL,
		created_at TIMESTAMP NOT NULL,
		processing_ms INTEGER NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_created_at ON analyses(created_at);
	CREATE INDEX IF NOT EXISTS idx_confidence ON analyses(confidence);
	`
	
	_, err := db.conn.Exec(query)
	return err
}

func (db *DB) SaveAnalysis(analysis *models.TextAnalysis) error {
	metadataJSON, err := json.Marshal(analysis.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	query := `
		INSERT INTO analyses (id, text, summary, metadata, confidence, created_at, processing_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err = db.conn.Exec(
		query,
		analysis.ID,
		analysis.Text,
		analysis.Summary,
		string(metadataJSON),
		analysis.Confidence,
		analysis.CreatedAt,
		analysis.ProcessingMS,
	)
	
	if err != nil {
		return fmt.Errorf("failed to insert analysis: %w", err)
	}
	
	return nil
}

func (db *DB) GetAnalysis(id string) (*models.TextAnalysis, error) {
	query := `
		SELECT id, text, summary, metadata, confidence, created_at, processing_ms
		FROM analyses
		WHERE id = ?
	`
	
	var analysis models.TextAnalysis
	var metadataJSON string
	
	err := db.conn.QueryRow(query, id).Scan(
		&analysis.ID,
		&analysis.Text,
		&analysis.Summary,
		&metadataJSON,
		&analysis.Confidence,
		&analysis.CreatedAt,
		&analysis.ProcessingMS,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query analysis: %w", err)
	}
	
	if err := json.Unmarshal([]byte(metadataJSON), &analysis.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	
	return &analysis, nil
}

func (db *DB) SearchAnalyses(query models.SearchQuery) ([]*models.TextAnalysis, error) {
	var conditions []string
	var args []interface{}
	
	baseQuery := `
		SELECT id, text, summary, metadata, confidence, created_at, processing_ms
		FROM analyses
		WHERE 1=1
	`
	
	if query.Topic != "" {
		conditions = append(conditions, "metadata LIKE ?")
		args = append(args, "%\""+query.Topic+"\"%")
	}
	
	if query.Keyword != "" {
		conditions = append(conditions, "(text LIKE ? OR summary LIKE ? OR metadata LIKE ?)")
		keyword := "%" + query.Keyword + "%"
		args = append(args, keyword, keyword, keyword)
	}
	
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}
	
	baseQuery += " ORDER BY created_at DESC"
	
	if query.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, query.Limit)
	}
	
	if query.Offset > 0 {
		baseQuery += " OFFSET ?"
		args = append(args, query.Offset)
	}
	
	rows, err := db.conn.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search analyses: %w", err)
	}
	defer rows.Close()
	
	var results []*models.TextAnalysis
	
	for rows.Next() {
		var analysis models.TextAnalysis
		var metadataJSON string
		
		err := rows.Scan(
			&analysis.ID,
			&analysis.Text,
			&analysis.Summary,
			&metadataJSON,
			&analysis.Confidence,
			&analysis.CreatedAt,
			&analysis.ProcessingMS,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		if err := json.Unmarshal([]byte(metadataJSON), &analysis.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		
		results = append(results, &analysis)
	}
	
	return results, nil
}

func (db *DB) GetRecentAnalyses(limit int) ([]*models.TextAnalysis, error) {
	query := models.SearchQuery{
		Limit: limit,
	}
	return db.SearchAnalyses(query)
}

func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM analyses").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["total_analyses"] = count
	
	var avgConfidence sql.NullFloat64
	err = db.conn.QueryRow("SELECT AVG(confidence) FROM analyses").Scan(&avgConfidence)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if avgConfidence.Valid {
		stats["average_confidence"] = avgConfidence.Float64
	} else {
		stats["average_confidence"] = 0.0
	}
	
	var avgProcessingTime sql.NullFloat64
	err = db.conn.QueryRow("SELECT AVG(processing_ms) FROM analyses").Scan(&avgProcessingTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if avgProcessingTime.Valid {
		stats["average_processing_ms"] = avgProcessingTime.Float64
	} else {
		stats["average_processing_ms"] = 0.0
	}
	
	var lastAnalysisStr sql.NullString
	err = db.conn.QueryRow("SELECT MAX(created_at) FROM analyses").Scan(&lastAnalysisStr)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if lastAnalysisStr.Valid && lastAnalysisStr.String != "" {
		stats["last_analysis"] = lastAnalysisStr.String
	}
	
	return stats, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}