package logging

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	once   sync.Once
)

// Config represents the logger configuration
type Config struct {
	Level      string
	Format     string // json or console
	OutputPath string
}

// InitLogger initializes the global logger
func InitLogger(config *Config) error {
	var err error
	once.Do(func() {
		// Parse log level
		level := zapcore.InfoLevel
		if err := level.UnmarshalText([]byte(config.Level)); err != nil {
			fmt.Printf("Invalid log level %s, defaulting to info\n", config.Level)
		}

		// Configure output
		outputPath := config.OutputPath
		if outputPath == "" {
			outputPath = "stdout"
		}

		// Create encoder config
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// Create encoder based on format
		var encoder zapcore.Encoder
		if config.Format == "json" {
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		// Configure output sink
		var sink zapcore.WriteSyncer
		if outputPath == "stdout" {
			sink = zapcore.AddSync(os.Stdout)
		} else if outputPath == "stderr" {
			sink = zapcore.AddSync(os.Stderr)
		} else {
			// Create log file
			logFile, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Failed to open log file: %v, defaulting to stdout\n", err)
				sink = zapcore.AddSync(os.Stdout)
			} else {
				sink = zapcore.AddSync(logFile)
			}
		}

		// Create the core
		core := zapcore.NewCore(encoder, sink, zap.NewAtomicLevelAt(level))

		// Create the logger
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		sugar = logger.Sugar()

		// Ensure logs are flushed on exit
		zap.RedirectStdLog(logger)
	})

	return err
}

// GetLogger returns the global zap logger
func GetLogger() *zap.Logger {
	if logger == nil {
		// Default initialization if not already initialized
		defaultConfig := &Config{
			Level:  "info",
			Format: "console",
		}
		if err := InitLogger(defaultConfig); err != nil {
			fmt.Printf("Failed to initialize default logger: %v\n", err)
		}
	}
	return logger
}

// GetSugaredLogger returns the global sugared logger
func GetSugaredLogger() *zap.SugaredLogger {
	if sugar == nil {
		// Default initialization if not already initialized
		defaultConfig := &Config{
			Level:  "info",
			Format: "console",
		}
		if err := InitLogger(defaultConfig); err != nil {
			fmt.Printf("Failed to initialize default logger: %v\n", err)
		}
	}
	return sugar
}

// With returns a logger with the given fields
func With(fields ...zapcore.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// WithFields returns a sugared logger with the given fields
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	return GetSugaredLogger().With(fields)
}

// Info logs an info message
func Info(msg string, fields ...zapcore.Field) {
	GetLogger().Info(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zapcore.Field) {
	GetLogger().Debug(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zapcore.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zapcore.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message and then exits
func Fatal(msg string, fields ...zapcore.Field) {
	GetLogger().Fatal(msg, fields...)
}

// IsDevelopment returns true if in development mode
func IsDevelopment() bool {
	// This is a simple implementation - you might want to use environment variables
	// or configuration settings to determine the environment
	return os.Getenv("GO_ENV") != "production"
}

// Sync flushes any buffered log entries
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
