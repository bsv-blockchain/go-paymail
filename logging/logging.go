package logging

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
	"go.elastic.co/ecszerolog"
)

var (
	defaultLogger     *zerolog.Logger
	defaultLoggerOnce sync.Once
)

// GetDefaultLogger generates and returns a default logger instance.
// Uses sync.Once to avoid race conditions in ecszerolog.New().
func GetDefaultLogger() *zerolog.Logger {
	defaultLoggerOnce.Do(func() {
		logger := ecszerolog.New(os.Stdout, ecszerolog.Level(zerolog.DebugLevel)).
			With().
			Timestamp().
			Caller().
			Str("application", "go-paymail").
			Logger()
		defaultLogger = &logger
	})

	return defaultLogger
}
