package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	// Default configuration
	cfg := getDefaultConfig()

	// Load from config file
	configFile := getConfigFile()
	if configFile != "" {
		if err := loadFromFile(cfg, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(cfg)

	// Validate configuration
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// getDefaultConfig returns default configuration values
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Worker: WorkerConfig{
			PoolSize:   4,
			QueueSize:  100,
			Timeout:    5 * time.Minute,
			RetryCount: 3,
			RetryDelay: 10 * time.Second,
		},
		Print: PrintConfig{
			MaxFileSize:     10 * 1024 * 1024, // 10MB
			OutputDirectory: "./output",
			TempDirectory:   "./temp",
			Timeout:         2 * time.Minute,
			MaxConcurrent:   2,
		},
		Queue: QueueConfig{
			Type:       "memory",
			MaxRetries: 3,
			Timeout:    30 * time.Second,
		},
		Cache: CacheConfig{
			Type:       "memory",
			TTL:        1 * time.Hour,
			MaxSize:    100 * 1024 * 1024, // 100MB
			MaxEntries: 1000,
		},
		Logger: LoggerConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100, // MB
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		},
	}
}

// getConfigFile determines which config file to use
func getConfigFile() string {
	// Check environment variable first
	if configFile := os.Getenv("CONFIG_FILE"); configFile != "" {
		return configFile
	}

	// Check for environment-specific config files
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	configPaths := []string{
		fmt.Sprintf("configs/%s.yaml", env),
		fmt.Sprintf("configs/%s.yml", env),
		"config.yaml",
		"config.yml",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(cfg *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(cfg *Config) {
	// Server configuration
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p := parseInt(port); p > 0 {
			cfg.Server.Port = p
		}
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	// Worker configuration
	if poolSize := os.Getenv("WORKER_POOL_SIZE"); poolSize != "" {
		if p := parseInt(poolSize); p > 0 {
			cfg.Worker.PoolSize = p
		}
	}

	// Print configuration
	if maxFileSize := os.Getenv("PRINT_MAX_FILE_SIZE"); maxFileSize != "" {
		if s := parseInt64(maxFileSize); s > 0 {
			cfg.Print.MaxFileSize = s
		}
	}
	if outputDir := os.Getenv("PRINT_OUTPUT_DIRECTORY"); outputDir != "" {
		cfg.Print.OutputDirectory = outputDir
	}
	if tempDir := os.Getenv("PRINT_TEMP_DIRECTORY"); tempDir != "" {
		cfg.Print.TempDirectory = tempDir
	}

	// Queue configuration
	if queueType := os.Getenv("QUEUE_TYPE"); queueType != "" {
		cfg.Queue.Type = queueType
	}
	if queueURL := os.Getenv("QUEUE_URL"); queueURL != "" {
		cfg.Queue.URL = queueURL
	}

	// Cache configuration
	if cacheType := os.Getenv("CACHE_TYPE"); cacheType != "" {
		cfg.Cache.Type = cacheType
	}
	if cacheURL := os.Getenv("CACHE_URL"); cacheURL != "" {
		cfg.Cache.URL = cacheURL
	}

	// Logger configuration
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logger.Level = strings.ToLower(logLevel)
	}
	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		cfg.Logger.Format = strings.ToLower(logFormat)
	}
}

// validate validates the configuration
func validate(cfg *Config) error {
	// Validate server configuration
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// Validate worker configuration
	if cfg.Worker.PoolSize <= 0 {
		return fmt.Errorf("worker pool size must be positive: %d", cfg.Worker.PoolSize)
	}

	// Validate print configuration
	if cfg.Print.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be positive: %d", cfg.Print.MaxFileSize)
	}

	// Create directories if they don't exist
	dirs := []string{cfg.Print.OutputDirectory, cfg.Print.TempDirectory}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Validate logger configuration
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true, "fatal": true,
	}
	if !validLogLevels[cfg.Logger.Level] {
		return fmt.Errorf("invalid log level: %s", cfg.Logger.Level)
	}

	validLogFormats := map[string]bool{
		"json": true, "text": true,
	}
	if !validLogFormats[cfg.Logger.Format] {
		return fmt.Errorf("invalid log format: %s", cfg.Logger.Format)
	}

	return nil
}

// Helper functions for parsing environment variables
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

func parseInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}

// GetConfigPath returns the absolute path to a config file
func GetConfigPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	
	// Look in configs directory first
	configsPath := filepath.Join("configs", filename)
	if _, err := os.Stat(configsPath); err == nil {
		abs, _ := filepath.Abs(configsPath)
		return abs
	}
	
	// Fall back to current directory
	abs, _ := filepath.Abs(filename)
	return abs
}
