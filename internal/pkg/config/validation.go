package config

import (
	"fmt"
	"os"
	"time"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("config validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%d configuration validation errors: %s (and %d more)", len(e), e[0].Error(), len(e)-1)
}

// Validate validates the entire configuration
func (c *Config) Validate() error {
	var errors ValidationErrors

	// Validate server configuration
	if err := c.validateServer(); err != nil {
		if validationErrs, ok := err.(ValidationErrors); ok {
			errors = append(errors, validationErrs...)
		} else {
			errors = append(errors, ValidationError{Field: "server", Message: err.Error()})
		}
	}

	// Validate print configuration
	if err := c.validatePrint(); err != nil {
		if validationErrs, ok := err.(ValidationErrors); ok {
			errors = append(errors, validationErrs...)
		} else {
			errors = append(errors, ValidationError{Field: "print", Message: err.Error()})
		}
	}

	// Validate logger configuration
	if err := c.validateLogger(); err != nil {
		if validationErrs, ok := err.(ValidationErrors); ok {
			errors = append(errors, validationErrs...)
		} else {
			errors = append(errors, ValidationError{Field: "logger", Message: err.Error()})
		}
	}

	// Validate queue configuration
	if err := c.validateQueue(); err != nil {
		if validationErrs, ok := err.(ValidationErrors); ok {
			errors = append(errors, validationErrs...)
		} else {
			errors = append(errors, ValidationError{Field: "queue", Message: err.Error()})
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateServer validates server configuration
func (c *Config) validateServer() error {
	var errors ValidationErrors

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors = append(errors, ValidationError{
			Field:   "server.port",
			Message: "port must be between 1 and 65535",
		})
	}

	if c.Server.ReadTimeout <= 0 {
		errors = append(errors, ValidationError{
			Field:   "server.read_timeout",
			Message: "read timeout must be positive",
		})
	}

	if c.Server.WriteTimeout <= 0 {
		errors = append(errors, ValidationError{
			Field:   "server.write_timeout",
			Message: "write timeout must be positive",
		})
	}

	if c.Server.IdleTimeout <= 0 {
		errors = append(errors, ValidationError{
			Field:   "server.idle_timeout",
			Message: "idle timeout must be positive",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validatePrint validates print configuration
func (c *Config) validatePrint() error {
	var errors ValidationErrors

	if c.Print.OutputDirectory == "" {
		errors = append(errors, ValidationError{
			Field:   "print.output_directory",
			Message: "output directory cannot be empty",
		})
	} else {
		// Check if directory exists and is writable
		if info, err := os.Stat(c.Print.OutputDirectory); err != nil {
			if os.IsNotExist(err) {
				// Try to create the directory
				if err := os.MkdirAll(c.Print.OutputDirectory, 0755); err != nil {
					errors = append(errors, ValidationError{
						Field:   "print.output_directory",
						Message: fmt.Sprintf("cannot create output directory: %v", err),
					})
				}
			} else {
				errors = append(errors, ValidationError{
					Field:   "print.output_directory",
					Message: fmt.Sprintf("cannot access output directory: %v", err),
				})
			}
		} else if !info.IsDir() {
			errors = append(errors, ValidationError{
				Field:   "print.output_directory",
				Message: "output directory path is not a directory",
			})
		}
	}

	if c.Print.MaxFileSize <= 0 {
		errors = append(errors, ValidationError{
			Field:   "print.max_file_size",
			Message: "max file size must be positive",
		})
	}

	if c.Print.Timeout <= 0 {
		errors = append(errors, ValidationError{
			Field:   "print.timeout",
			Message: "timeout must be positive",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateLogger validates logger configuration
func (c *Config) validateLogger() error {
	var errors ValidationErrors

	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	if !validLevels[c.Logger.Level] {
		errors = append(errors, ValidationError{
			Field:   "logger.level",
			Message: "level must be one of: debug, info, warn, error, fatal",
		})
	}

	validOutputs := map[string]bool{
		"stdout": true,
		"stderr": true,
		"file":   true,
	}

	if !validOutputs[c.Logger.Output] {
		errors = append(errors, ValidationError{
			Field:   "logger.output",
			Message: "output must be one of: stdout, stderr, file",
		})
	}

	if c.Logger.Output == "file" && c.Logger.File == "" {
		errors = append(errors, ValidationError{
			Field:   "logger.file",
			Message: "file path is required when output is 'file'",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// validateQueue validates queue configuration
func (c *Config) validateQueue() error {
	var errors ValidationErrors

	if c.Queue.MaxRetries < 0 {
		errors = append(errors, ValidationError{
			Field:   "queue.max_retries",
			Message: "max retries cannot be negative",
		})
	}

	if c.Queue.Timeout <= 0 {
		errors = append(errors, ValidationError{
			Field:   "queue.timeout",
			Message: "timeout must be positive",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// SetDefaults sets default values for missing configuration
func (c *Config) SetDefaults() {
	// Server defaults
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 60 * time.Second
	}

	// Print defaults
	if c.Print.OutputDirectory == "" {
		c.Print.OutputDirectory = "./output"
	}
	if c.Print.MaxFileSize == 0 {
		c.Print.MaxFileSize = 50 * 1024 * 1024 // 50MB
	}
	if c.Print.Timeout == 0 {
		c.Print.Timeout = 5 * time.Minute
	}

	// Logger defaults
	if c.Logger.Level == "" {
		c.Logger.Level = "info"
	}
	if c.Logger.Output == "" {
		c.Logger.Output = "stdout"
	}

	// Worker defaults (using Worker config instead of Queue for worker-specific settings)
	if c.Worker.PoolSize == 0 {
		c.Worker.PoolSize = 4
	}
	if c.Worker.QueueSize == 0 {
		c.Worker.QueueSize = 100
	}
	if c.Worker.RetryCount == 0 {
		c.Worker.RetryCount = 3
	}
	if c.Worker.RetryDelay == 0 {
		c.Worker.RetryDelay = 30 * time.Second
	}
	if c.Worker.Timeout == 0 {
		c.Worker.Timeout = 10 * time.Minute
	}

	// Queue defaults
	if c.Queue.MaxRetries == 0 {
		c.Queue.MaxRetries = 3
	}
	if c.Queue.Timeout == 0 {
		c.Queue.Timeout = 10 * time.Minute
	}

	// Cache defaults
	if c.Cache.TTL == 0 {
		c.Cache.TTL = 1 * time.Hour
	}
}
