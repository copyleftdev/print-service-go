package logger

import (
	"os"

	"print-service/internal/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// StructuredLogger implements Logger using zap
type StructuredLogger struct {
	logger *zap.SugaredLogger
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(cfg *config.LoggerConfig) Logger {
	// Configure log level
	level := ParseLogLevel(cfg.Level)
	// Safe conversion to avoid integer overflow
	var zapLevel zapcore.Level
	if level >= -128 && level <= 127 {
		zapLevel = zapcore.Level(level)
	} else {
		zapLevel = zapcore.InfoLevel // Default to info level if out of range
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Configure encoder
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	case "file":
		if cfg.File != "" {
			file, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err == nil {
				writeSyncer = zapcore.AddSync(file)
			} else {
				writeSyncer = zapcore.AddSync(os.Stdout)
			}
		} else {
			writeSyncer = zapcore.AddSync(os.Stdout)
		}
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &StructuredLogger{
		logger: logger.Sugar(),
	}
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debugw(msg, keysAndValues...)
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Infow(msg, keysAndValues...)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warnw(msg, keysAndValues...)
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Errorw(msg, keysAndValues...)
}

// Fatal logs a fatal message and exits
func (l *StructuredLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.logger.Fatalw(msg, keysAndValues...)
}

// With returns a logger with additional context
func (l *StructuredLogger) With(keysAndValues ...interface{}) Logger {
	return &StructuredLogger{
		logger: l.logger.With(keysAndValues...),
	}
}

// Sync flushes any buffered log entries
func (l *StructuredLogger) Sync() error {
	return l.logger.Sync()
}
