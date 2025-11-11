package utils

import (
	"github.com/ibn-network/api-gateway/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new zap logger based on configuration
func NewLogger(cfg *config.LoggingConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	// Set log level
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	// Configure based on format
	if cfg.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Set output
	if cfg.Output == "stderr" {
		zapConfig.OutputPaths = []string{"stderr"}
	} else if cfg.Output == "file" {
		zapConfig.OutputPaths = []string{"logs/app.log"}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
	}

	// Build logger
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// NewDevelopmentLogger creates a development logger
func NewDevelopmentLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

// NewProductionLogger creates a production logger
func NewProductionLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

