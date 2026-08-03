package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is the application-wide logger. Everything (startup, config,
// repository lifecycle, and request errors) is written through this logger.
var Log = zerolog.New(os.Stdout).With().Timestamp().Logger()

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
}
