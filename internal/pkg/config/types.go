package config

import (
	"time"
)

// Config represents the application configuration
type Config struct {
	Server ServerConfig `yaml:"server" json:"server"`
	Worker WorkerConfig `yaml:"worker" json:"worker"`
	Print  PrintConfig  `yaml:"print" json:"print"`
	Queue  QueueConfig  `yaml:"queue" json:"queue"`
	Cache  CacheConfig  `yaml:"cache" json:"cache"`
	Logger LoggerConfig `yaml:"logger" json:"logger"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port         int           `yaml:"port" json:"port"`
	Host         string        `yaml:"host" json:"host"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	TLS          TLSConfig     `yaml:"tls" json:"tls"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

// WorkerConfig represents background worker configuration
type WorkerConfig struct {
	PoolSize   int           `yaml:"pool_size" json:"pool_size"`
	QueueSize  int           `yaml:"queue_size" json:"queue_size"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
	RetryCount int           `yaml:"retry_count" json:"retry_count"`
	RetryDelay time.Duration `yaml:"retry_delay" json:"retry_delay"`
}

// PrintConfig represents print service configuration
type PrintConfig struct {
	MaxFileSize     int64         `yaml:"max_file_size" json:"max_file_size"`
	OutputDirectory string        `yaml:"output_directory" json:"output_directory"`
	TempDirectory   string        `yaml:"temp_directory" json:"temp_directory"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	MaxConcurrent   int           `yaml:"max_concurrent" json:"max_concurrent"`
}

// QueueConfig represents queue configuration
type QueueConfig struct {
	Type       string        `yaml:"type" json:"type"` // memory, redis, etc.
	URL        string        `yaml:"url" json:"url"`
	MaxRetries int           `yaml:"max_retries" json:"max_retries"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	Type       string        `yaml:"type" json:"type"` // memory, redis, etc.
	URL        string        `yaml:"url" json:"url"`
	TTL        time.Duration `yaml:"ttl" json:"ttl"`
	MaxSize    int64         `yaml:"max_size" json:"max_size"`
	MaxEntries int           `yaml:"max_entries" json:"max_entries"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Level      string `yaml:"level" json:"level"`
	Format     string `yaml:"format" json:"format"` // json, text
	Output     string `yaml:"output" json:"output"` // stdout, stderr, file
	File       string `yaml:"file" json:"file"`
	MaxSize    int    `yaml:"max_size" json:"max_size"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups"`
	MaxAge     int    `yaml:"max_age" json:"max_age"`
	Compress   bool   `yaml:"compress" json:"compress"`
}
