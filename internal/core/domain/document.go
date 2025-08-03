package domain

import (
	"time"
)

// Document represents a document to be printed
type Document struct {
	ID          string           `json:"id"`
	Content     string           `json:"content"`
	ContentType ContentType      `json:"content_type"`
	Metadata    DocumentMetadata `json:"metadata"`
	Options     PrintOptions     `json:"options"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// ContentType represents the type of document content
type ContentType string

const (
	ContentTypeHTML     ContentType = "html"
	ContentTypeMarkdown ContentType = "markdown"
	ContentTypeText     ContentType = "text"
)

// DocumentMetadata contains metadata about the document
type DocumentMetadata struct {
	Title       string            `json:"title"`
	Author      string            `json:"author"`
	Subject     string            `json:"subject"`
	Keywords    []string          `json:"keywords"`
	Creator     string            `json:"creator"`
	Producer    string            `json:"producer"`
	CustomProps map[string]string `json:"custom_props"`
}

// PrintJob represents a print job in the system
type PrintJob struct {
	ID          string      `json:"id"`
	Document    Document    `json:"document"`
	Status      JobStatus   `json:"status"`
	Progress    float64     `json:"progress"`
	Error       string      `json:"error,omitempty"`
	OutputPath  string      `json:"output_path,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	StartedAt   *time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	RetryCount  int         `json:"retry_count"`
	Priority    JobPriority `json:"priority"`
}

// JobStatus represents the status of a print job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// JobPriority represents the priority of a print job
type JobPriority int

const (
	PriorityLow    JobPriority = 1
	PriorityNormal JobPriority = 5
	PriorityHigh   JobPriority = 10
)

// RenderResult represents the result of a document rendering operation
type RenderResult struct {
	OutputPath string        `json:"output_path"`
	OutputSize int64         `json:"output_size"`
	PageCount  int           `json:"page_count"`
	RenderTime time.Duration `json:"render_time"`
	CacheHit   bool          `json:"cache_hit"`
	Warnings   []string      `json:"warnings,omitempty"`
}
