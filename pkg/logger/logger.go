package logger

import (
	"context"

	"cloud.google.com/go/logging"
)

// Instead of Global
// https://pkg.go.dev/golang.org/x/tools/internal/memoize?tab=doc

type Logger struct {
	*logging.Logger
	*logging.Client
}

var logger *Logger

func CreateLogger(ctx context.Context, projectID, loggerName string) *Logger {
	logClient, err := logging.NewClient(ctx, projectID)
	if err != nil {

	}
	return &Logger{
		Logger: logClient.Logger(loggerName),
		Client: logClient,
	}
}

// Lock->Log->Unlock
func Log(message string) {
	logger.Logger.Log(
		logging.Entry{Payload: message},
	)
}

func Error(message string) {
	logger.Logger.Log(
		logging.Entry{
			Payload:  message,
			Severity: logging.Error,
		},
	)
}
