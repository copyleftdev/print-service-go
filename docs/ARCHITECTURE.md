# Architecture Documentation

## Overview

The print service is designed as a modular, scalable system for converting HTML documents to PDF format. It follows clean architecture principles with clear separation of concerns.

## System Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │    │   Background    │
│                 │    │     Worker      │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          ▼                      ▼
┌─────────────────────────────────────────┐
│              Load Balancer              │
└─────────────────┬───────────────────────┘
                  │
          ┌───────┴───────┐
          ▼               ▼
┌─────────────────┐ ┌─────────────────┐
│  HTTP Server    │ │  Worker Pool    │
│                 │ │                 │
│ • REST API      │ │ • Job Queue     │
│ • Health Check  │ │ • Processing    │
│ • Rate Limiting │ │ • Concurrency   │
└─────────┬───────┘ └─────────┬───────┘
          │                   │
          └───────┬───────────┘
                  │
          ┌───────▼───────┐
          │  Core Engine  │
          │               │
          │ • HTML Parser │
          │ • CSS Engine  │
          │ • Layout Calc │
          │ • PDF Gen     │
          └───────┬───────┘
                  │
    ┌─────────────┼─────────────┐
    ▼             ▼             ▼
┌─────────┐ ┌─────────┐ ┌─────────┐
│  Cache  │ │ Storage │ │ Logging │
│         │ │         │ │         │
│ • Redis │ │ • Files │ │ • Zap   │
│ • Memory│ │ • S3    │ │ • Stdout│
└─────────┘ └─────────┘ └─────────┘
```

## Directory Structure

```
print-service/
├── cmd/                    # Application entry points
│   ├── server/            # HTTP server binary
│   └── worker/            # Background worker binary
├── internal/              # Private application code
│   ├── api/               # HTTP API layer
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── middleware/    # HTTP middleware
│   │   └── routes/        # Route definitions
│   ├── core/              # Business logic layer
│   │   ├── domain/        # Domain models and errors
│   │   ├── engine/        # Processing engines
│   │   │   ├── html/      # HTML parsing and sanitization
│   │   │   ├── css/       # CSS parsing and styling
│   │   │   ├── layout/    # Layout calculation
│   │   │   └── render/    # PDF rendering
│   │   └── services/      # Core business services
│   ├── infrastructure/    # External dependencies
│   │   ├── cache/         # Caching implementations
│   │   ├── storage/       # File storage
│   │   ├── queue/         # Job queue
│   │   └── logger/        # Logging infrastructure
│   ├── pkg/               # Shared utilities
│   │   ├── config/        # Configuration management
│   │   ├── pool/          # Worker pool implementation
│   │   └── utils/         # Common utilities
│   └── tests/             # Test suites
│       └── golden/        # Golden test framework
├── configs/               # Configuration files
├── assets/                # Static assets
├── docs/                  # Documentation
└── testdata/              # Test data
```

## Core Components

### HTML Engine

**Location**: `internal/core/engine/html/`

- **Parser**: Converts HTML strings to DOM tree structures
- **Sanitizer**: Removes dangerous content and validates HTML structure
- **Validator**: Ensures HTML integrity and compliance

**Key Features**:
- Cycle detection to prevent infinite recursion
- Depth limiting for complex nested structures
- XSS protection through content sanitization
- Support for malformed HTML graceful handling

### CSS Engine

**Location**: `internal/core/engine/css/`

- **Parser**: Parses CSS rules and selectors
- **Selector**: Matches CSS rules to DOM elements
- **Cascade**: Applies CSS cascade and inheritance rules

**Key Features**:
- Complete CSS selector support
- Cascade and specificity calculation
- Inheritance and computed value resolution
- Box model implementation

### Layout Engine

**Location**: `internal/core/engine/layout/`

- **Box Calculator**: Implements CSS box model
- **Flow Calculator**: Handles text flow and line breaking
- **Position Calculator**: Manages element positioning

**Key Features**:
- CSS box model (margin, border, padding, content)
- Text flow and line breaking algorithms
- Absolute and relative positioning
- Float and clear handling

### Render Engine

**Location**: `internal/core/engine/render/`

- **PDF Generator**: Converts layout to PDF format
- **Font Manager**: Handles font loading and metrics
- **Image Processor**: Processes and embeds images

**Key Features**:
- High-quality PDF output
- Font embedding and subsetting
- Image optimization and embedding
- Vector graphics support

## Data Flow

### Request Processing

1. **HTTP Request**: Client submits HTML via REST API
2. **Validation**: Input validation and sanitization
3. **Queue**: Job queued for background processing
4. **Processing**: Worker picks up job and processes
5. **Response**: Result stored and download link provided

### Processing Pipeline

```
HTML Input
    ↓
HTML Parser → DOM Tree
    ↓
CSS Parser → Style Rules
    ↓
Style Calculator → Styled DOM
    ↓
Layout Calculator → Layout Tree
    ↓
PDF Renderer → PDF Output
```

## Concurrency Model

### HTTP Server

- **Goroutine per request**: Each HTTP request handled in separate goroutine
- **Connection pooling**: Database and cache connections pooled
- **Rate limiting**: Request rate limiting to prevent abuse
- **Graceful shutdown**: Clean shutdown with request completion

### Worker Pool

- **Fixed pool size**: Configurable number of worker goroutines
- **Job queue**: Channel-based job distribution
- **Load balancing**: Jobs distributed evenly across workers
- **Error handling**: Failed jobs retry with exponential backoff

### Synchronization

- **Mutex protection**: Shared resources protected with mutexes
- **Channel communication**: Goroutines communicate via channels
- **Context cancellation**: Request cancellation propagated through context
- **Wait groups**: Coordinated shutdown of goroutine groups

## Error Handling

### Error Types

- **Validation Errors**: Input validation failures
- **Processing Errors**: HTML/CSS parsing or layout errors
- **System Errors**: Infrastructure failures (cache, storage)
- **Timeout Errors**: Processing timeout exceeded

### Error Recovery

- **Graceful degradation**: Fallback to simplified processing
- **Retry mechanisms**: Automatic retry with exponential backoff
- **Circuit breakers**: Prevent cascade failures
- **Health monitoring**: Continuous health checks

## Security Considerations

### Input Sanitization

- **HTML sanitization**: Remove dangerous HTML elements and attributes
- **CSS validation**: Validate CSS properties and values
- **XSS prevention**: Content Security Policy and input escaping
- **File upload limits**: Size and type restrictions

### Access Control

- **Rate limiting**: Prevent abuse and DoS attacks
- **Input validation**: Strict validation of all inputs
- **Error information**: Limited error details in responses
- **Logging**: Comprehensive security event logging

## Performance Optimization

### Caching Strategy

- **Template caching**: Frequently used templates cached
- **Font caching**: Font metrics and data cached
- **Image caching**: Processed images cached
- **Result caching**: Generated PDFs cached with TTL

### Memory Management

- **Object pooling**: Reuse of expensive objects
- **Garbage collection**: Optimized GC settings
- **Memory limits**: Configurable memory limits per job
- **Resource cleanup**: Proper cleanup of resources

### I/O Optimization

- **Async processing**: Non-blocking I/O operations
- **Batch operations**: Batch database and cache operations
- **Connection pooling**: Reuse of network connections
- **Compression**: Response compression for bandwidth efficiency

## Monitoring and Observability

### Metrics

- **Request metrics**: Request count, duration, status codes
- **Processing metrics**: Job queue size, processing time
- **System metrics**: Memory usage, CPU utilization, goroutine count
- **Error metrics**: Error rates and types

### Logging

- **Structured logging**: JSON-formatted logs with consistent fields
- **Log levels**: Debug, info, warn, error levels
- **Request tracing**: Request ID tracking through pipeline
- **Performance logging**: Processing time and resource usage

### Health Checks

- **Liveness probe**: Service is running and responsive
- **Readiness probe**: Service is ready to accept requests
- **Dependency checks**: External dependency health validation
- **Resource checks**: Memory and disk usage validation

## Deployment Architecture

### Container Strategy

- **Multi-stage builds**: Optimized Docker images
- **Security scanning**: Container vulnerability scanning
- **Resource limits**: CPU and memory limits configured
- **Health checks**: Container health check endpoints

### Scaling Strategy

- **Horizontal scaling**: Multiple service instances
- **Load balancing**: Request distribution across instances
- **Auto-scaling**: Automatic scaling based on metrics
- **Resource allocation**: CPU and memory resource management

### Configuration Management

- **Environment-based**: Different configs for different environments
- **Secret management**: Secure handling of sensitive configuration
- **Hot reload**: Configuration changes without restart
- **Validation**: Configuration validation at startup
