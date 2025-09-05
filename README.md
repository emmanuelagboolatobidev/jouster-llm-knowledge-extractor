# Jouster LLM Knowledge Extractor

A Go application that extracts structured knowledge from unstructured text using LLM

## Design Choices & Architecture Decisions

 I implemented a **provider interface pattern** for LLM integration, allowing seamless switching between multiple llm implementations without changing core logic. **SQLite** was selected for its zero-configuration deployment and portability, while still providing robust querying capabilities through JSON field searches. The local keyword extraction using suffix-based noun detection avoids unnecessary LLM calls, reducing costs and latency for a frequently-needed feature.

## Trade-offs Due to Time Constraints

Given the time limit, I prioritized **core functionality over optimization**: the mock provider i used returns static topics rather than contextual analysis, and the confidence scoring uses simple heuristics instead of ML-based evaluation. I chose **SQLite over PostgreSQL** for faster setup despite losing advanced full-text search capabilities and concurrent write performance. Error handling is functional but could be more granular. currently all LLM failures return the same error code rather than distinguishing between rate limits, network issues, and invalid responses. Finally, I skipped implementing authentication, rate limiting, and caching layers that would be essential for production deployment.

## Features

- **Text Analysis**: Generate summaries and extract structured metadata
- **Keyword Extraction**: Identify top 3 most frequent nouns (implemented locally, not via LLM)
- **Multiple LLM Providers**: Support for other llm such as OpenAI, Claude, or Mock provider
- **Batch Processing**: Analyze multiple texts concurrently
- **SQLite Persistence**: Store all analyses with search capabilities
- **REST API**: Clean API endpoints for analysis and search
- **Confidence Scoring**: Heuristic-based confidence calculation
- **Error Handling**: Graceful handling of empty inputs and LLM failures
- **Docker Support**: Containerized deployment

## API Endpoints

### POST /analyze
Analyze a single text and store the result.

```bash
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Your text content here..."
  }'
```

Response:
```json
{
  "id": "uuid",
  "summary": "1-2 sentence summary",
  "metadata": {
    "title": "Extracted title",
    "topics": ["topic1", "topic2", "topic3"],
    "sentiment": "positive",
    "keywords": ["keyword1", "keyword2", "keyword3"]
  },
  "confidence": 0.85
}
```

### POST /batch-analyze
Analyze multiple texts (max 10 per batch).

```bash
curl -X POST http://localhost:8080/batch-analyze \
  -H "Content-Type: application/json" \
  -d '{
    "texts": ["Text 1...", "Text 2...", "Text 3..."]
  }'
```

### GET /search
Search stored analyses by topic or keyword.

```bash
curl "http://localhost:8080/search?topic=technology&limit=10"
curl "http://localhost:8080/search?keyword=innovation"
```

## Setup

### Prerequisites
- Go 1.21+
- SQLite3
- Docker (optional)
- Make (optional if you want to run scripts with make)

### Configuration

1. Copy the example environment file:
```bash
cp .env.example .env
```

### Running Locally

```bash
# Install dependencies
make deps

# Run tests
make test

# Run the application
make run
```

### Running with Docker

```bash
# Build and run with Docker Compose
make docker-run

# View logs
make docker-logs

# Stop containers
make docker-stop
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-verbose

# Run benchmarks
make bench
```

## Architecture

```
.
├── cmd/api/           # Application entry point
├── internal/
│   ├── analyzer/      # Keyword extraction logic
│   ├── database/      # SQLite persistence layer
│   ├── handlers/      # HTTP request handlers
│   ├── llm/          # LLM provider interfaces
│   └── models/       # Data structures
└── data/             # SQLite database storage
```

## Edge Cases Handled

1. **Empty Input**: Returns 400 error with clear message
2. **LLM API Failure**: Returns 503 with fallback behavior
3. **Concurrent Batch Processing**: Limited concurrency (3 parallel)
4. **Invalid JSON from LLM**: Applies defaults for missing fields
5. **Database Errors**: Proper error responses
6. **Context Timeouts**: 30-45 second timeouts with cancellation

## Performance

- Keyword extraction: ~1ms for typical text
- Batch processing: Concurrent with semaphore control
- Database queries: Indexed for performance

## Development

### Adding a New LLM Provider

1. Implement the `Provider` interface in `internal/llm/`
2. Add provider initialization in `NewProvider()`
3. Update configuration in `.env`

### Database Schema

```sql
CREATE TABLE analyses (
    id TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    summary TEXT NOT NULL,
    metadata TEXT NOT NULL,
    confidence REAL NOT NULL,
    created_at TIMESTAMP NOT NULL,
    processing_ms INTEGER NOT NULL
);
```