package logging

import (
	"context"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

var (
	logger *otelzap.Logger
)

// Init initializes the global logger. Call this early in main.
func Init() {
	z, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = otelzap.New(z)
	otelzap.ReplaceGlobals(logger)
}

// fallbackLogger returns a development logger if Init() was not called.
func fallbackLogger() *otelzap.Logger {
	z, _ := zap.NewDevelopment()
	return otelzap.New(z)
}

// L returns the global otelzap.Logger (for advanced use).
func L() *otelzap.Logger {
	if logger != nil {
		return logger
	}
	return fallbackLogger()
}

// C returns a context-aware logger (recommended for most use).
func C(ctx context.Context) otelzap.LoggerWithCtx {
	return L().Ctx(ctx)
}
