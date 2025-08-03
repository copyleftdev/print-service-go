# API Documentation

## Overview

The print service provides a RESTful HTTP API for converting HTML documents to PDF format. The API supports both synchronous and asynchronous processing modes.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication. This may change in future versions for production deployments.

## Health & Monitoring Endpoints

### Health Check

Check if the service is running and healthy.

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

### Readiness Check

Check if the service is ready to accept requests.

```http
GET /ready
```

**Response:**
```json
{
  "status": "ready",
  "timestamp": "2024-01-15T10:30:00Z",
  "dependencies": {
    "database": "healthy",
    "cache": "healthy"
  }
}
```

### Metrics

Get service metrics and statistics.

```http
GET /metrics
```

**Response:**
```json
{
  "requests_total": 1234,
  "requests_per_second": 12.5,
  "active_jobs": 3,
  "queue_size": 15,
  "memory_usage": "256MB",
  "uptime": "2h30m15s"
}
```

## Print Operations

### Submit Print Job

Submit an HTML document for PDF conversion.

```http
POST /api/v1/print
Content-Type: application/json
```

**Request Body:**
```json
{
  "html": "<html><body><h1>Hello World</h1></body></html>",
  "options": {
    "format": "A4",
    "orientation": "portrait",
    "margin": {
      "top": "1in",
      "bottom": "1in",
      "left": "1in",
      "right": "1in"
    },
    "header": {
      "enabled": true,
      "content": "Document Header"
    },
    "footer": {
      "enabled": true,
      "content": "Page {page} of {total}"
    }
  },
  "metadata": {
    "title": "My Document",
    "author": "John Doe",
    "subject": "Test Document"
  }
}
```

**Response:**
```json
{
  "job_id": "job_123456789",
  "status": "queued",
  "created_at": "2024-01-15T10:30:00Z",
  "estimated_completion": "2024-01-15T10:30:30Z"
}
```

### Get Job Status

Check the status of a print job.

```http
GET /api/v1/print/{job_id}
```

**Response:**
```json
{
  "job_id": "job_123456789",
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:30:25Z",
  "progress": 100,
  "result": {
    "pages": 3,
    "file_size": 245760,
    "download_url": "/api/v1/print/job_123456789/download"
  }
}
```

### Download Result

Download the generated PDF file.

```http
GET /api/v1/print/{job_id}/download
```

**Response:**
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename="document.pdf"`
- Binary PDF data

### Cancel Job

Cancel a pending or running print job.

```http
DELETE /api/v1/print/{job_id}
```

**Response:**
```json
{
  "job_id": "job_123456789",
  "status": "cancelled",
  "message": "Job cancelled successfully"
}
```

### List Jobs

Get a list of all print jobs.

```http
GET /api/v1/jobs?status=completed&limit=10&offset=0
```

**Query Parameters:**
- `status` (optional): Filter by job status (`queued`, `processing`, `completed`, `failed`, `cancelled`)
- `limit` (optional): Number of results to return (default: 50, max: 100)
- `offset` (optional): Number of results to skip (default: 0)

**Response:**
```json
{
  "jobs": [
    {
      "job_id": "job_123456789",
      "status": "completed",
      "created_at": "2024-01-15T10:30:00Z",
      "completed_at": "2024-01-15T10:30:25Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

## Request/Response Formats

### Job Status Values

- `queued`: Job is waiting to be processed
- `processing`: Job is currently being processed
- `completed`: Job completed successfully
- `failed`: Job failed with an error
- `cancelled`: Job was cancelled by user request

### Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "HTML content is required",
    "details": {
      "field": "html",
      "reason": "missing_required_field"
    }
  },
  "request_id": "req_123456789"
}
```

### Common Error Codes

- `INVALID_INPUT`: Request validation failed
- `JOB_NOT_FOUND`: Specified job ID does not exist
- `PROCESSING_ERROR`: Error occurred during PDF generation
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Unexpected server error

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Default Limit**: 100 requests per minute per IP
- **Burst Limit**: 10 requests per second
- **Headers**: Rate limit information is included in response headers

**Rate Limit Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248600
```

## Examples

### cURL Examples

Submit a simple print job:
```bash
curl -X POST http://localhost:8080/api/v1/print \
  -H "Content-Type: application/json" \
  -d '{
    "html": "<html><body><h1>Hello World</h1></body></html>",
    "options": {
      "format": "A4",
      "orientation": "portrait"
    }
  }'
```

Check job status:
```bash
curl http://localhost:8080/api/v1/print/job_123456789
```

Download result:
```bash
curl -o document.pdf http://localhost:8080/api/v1/print/job_123456789/download
```

### JavaScript Example

```javascript
// Submit print job
const response = await fetch('/api/v1/print', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    html: '<html><body><h1>Hello World</h1></body></html>',
    options: {
      format: 'A4',
      orientation: 'portrait'
    }
  })
});

const job = await response.json();
console.log('Job ID:', job.job_id);

// Poll for completion
const checkStatus = async (jobId) => {
  const statusResponse = await fetch(`/api/v1/print/${jobId}`);
  const status = await statusResponse.json();
  
  if (status.status === 'completed') {
    window.location.href = `/api/v1/print/${jobId}/download`;
  } else if (status.status === 'failed') {
    console.error('Job failed:', status.error);
  } else {
    setTimeout(() => checkStatus(jobId), 1000);
  }
};

checkStatus(job.job_id);
```

## WebSocket Support

For real-time job status updates, the service provides WebSocket endpoints:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/jobs/watch');

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Job update:', update);
};

// Subscribe to specific job
ws.send(JSON.stringify({
  action: 'subscribe',
  job_id: 'job_123456789'
}));
```
