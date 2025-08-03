# Pure Go Print Service

An HTML-to-PDF conversion service written in Go.

## Features

- HTML/CSS parsing with sanitization and validation
- CSS layout engine with box model and text flow calculations
- PDF generation using gofpdf library
- HTTP API server and background worker processes
- In-memory caching with configurable TTL
- HTML sanitization and input validation
- Health checks and structured logging
- Configurable worker pools and rate limiting

## Architecture

```
print-service/
├── cmd/
│   ├── server/          # HTTP server entry point
│   └── worker/          # Background worker entry point
├── internal/
│   ├── api/             # HTTP API layer
│   ├── core/            # Business logic
│   │   ├── domain/      # Domain types and errors
│   │   ├── engine/      # HTML/CSS/Layout engines
│   │   └── services/    # Core services
│   ├── infrastructure/ # External dependencies
│   └── pkg/            # Shared utilities
├── configs/            # Configuration files
└── assets/            # Static assets
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Optional: Redis for production caching/queuing

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd print-service-go
```

2. Install dependencies:
```bash
go mod download
```

3. Run the development server:
```bash
go run cmd/server/main.go
```

4. Run the background worker:
```bash
go run cmd/worker/main.go
```

### Usage

#### Print HTML to PDF

```bash
curl -X POST http://localhost:8080/api/v1/print \
  -H "Content-Type: application/json" \
  -d '{
    "content": "<html><body><h1>Hello World</h1></body></html>",
    "content_type": "html",
    "options": {
      "page": {
        "size": {"width": 210, "height": 297, "name": "A4"},
        "orientation": "portrait",
        "margins": {"top": 20, "right": 20, "bottom": 20, "left": 20}
      },
      "output": {
        "format": "pdf"
      }
    }
  }'
```

#### Check Job Status

```bash
curl http://localhost:8080/api/v1/print/{job_id}
```

#### Download Generated File

```bash
curl http://localhost:8080/api/v1/print/{job_id}/download -o output.pdf
```

## Configuration

The service uses YAML configuration files:

- `configs/development.yaml` - Development settings
- `configs/production.yaml` - Production settings

Key configuration sections:

- Server: HTTP server settings (port, timeouts, TLS)
- Worker: Background worker pool configuration
- Print: Document processing limits and paths
- Cache: Caching strategy and limits
- Logger: Logging configuration

## API Endpoints

### Health & Monitoring

- `GET /health` - Service health check
- `GET /ready` - Service readiness check
- `GET /metrics` - Service metrics

### Print Operations

- `POST /api/v1/print` - Submit print job
- `GET /api/v1/print/{id}` - Get job status
- `DELETE /api/v1/print/{id}` - Cancel job
- `GET /api/v1/print/{id}/download` - Download result
- `GET /api/v1/jobs` - List all jobs

## Core Components

### HTML Engine
- Parser: Converts HTML to DOM tree
- Sanitizer: Removes dangerous content for security
- Validator: Ensures HTML structure integrity

### CSS Engine
- Parser: Parses CSS rules and selectors
- Selector: Matches CSS rules to DOM elements
- Cascade: Applies CSS cascade and inheritance

### Layout Engine
- Box Calculator: Implements CSS box model
- Text Engine: Handles text layout and line breaking
- Flow Engine: Manages document flow (block, inline, flex)
- Page Breaker: Calculates optimal page breaks

### Services
- Print Service: Orchestrates the print pipeline
- Cache Service: Manages document and result caching
- Queue Service: Handles job queuing and processing
- Storage Service: Manages file storage operations

## Security

- HTML sanitization with configurable allowed tags/attributes
- Domain whitelist/blacklist for external resources
- Input validation and size limits
- Rate limiting for API endpoints
- Secure file handling

## Performance

- Caching with TTL
- Worker pools for concurrent processing
- Resource limits and timeouts
- Memory management and cleanup
- Rendering pipeline optimization

## Monitoring

- Structured JSON logging
- Metrics collection
- Health and readiness checks
- Request tracing and correlation IDs
- Performance monitoring

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
# Build server
go build -o bin/server cmd/server/main.go

# Build worker
go build -o bin/worker cmd/worker/main.go
```

### Docker

```bash
# Build image
docker build -t print-service .

# Run container
docker run -p 8080:8080 print-service
```

## Production Deployment

1. Configure production settings in `configs/production.yaml`
2. Set up Redis for caching and queuing
3. Configure TLS certificates
4. Set up monitoring and alerting
5. Deploy with container orchestration (Kubernetes, Docker Swarm)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue in the GitHub repository.
