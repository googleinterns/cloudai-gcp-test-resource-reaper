package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/logging"
)

// Instead of Global
// https://pkg.go.dev/golang.org/x/tools/internal/memoize?tab=doc

var logger *Logger

func CreateLogger() error {
	var err error
	logger, err = NewLogger()
	return err
}

func Log(v ...interface{}) {
	logger.Log(v...)
}

func Logf(format string, v ...interface{}) {
	logger.Logf(format, v...)
}

func Error(v ...interface{}) {
	fmt.Println(v)
	logger.Error(v...)
}

func Close() {
	if logger.CloudLogger != nil {
		logger.CloudLogger.Close()
	}
}

func AddCloudLogger(ctx context.Context, projectID, loggerName string) error {
	return logger.AddCloudLogger(ctx, projectID, loggerName)
}

type Logger struct {
	*CloudLogger
	*log.Logger
}

func NewLogger() (*Logger, error) {
	logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}
	fileLogger := log.New(logFile, "", log.Ldate|log.Ltime)
	return &Logger{Logger: fileLogger}, nil
}

func (l *Logger) AddCloudLogger(ctx context.Context, projectID, loggerName string) error {
	cloudLogger, err := CreateCloudLogger(ctx, projectID, loggerName)
	if err != nil {
		return err
	}
	l.CloudLogger = cloudLogger
	return nil
}

func (l *Logger) Log(v ...interface{}) {
	l.Logger.Println(v...)
	if l.CloudLogger != nil {
		l.CloudLogger.Log(v...)
	}
}

func (l *Logger) Logf(format string, v ...interface{}) {
	l.Logger.Printf(format, v...)
	if l.CloudLogger != nil {
		l.CloudLogger.Logf(format, v...)
	}
}

func (l *Logger) Error(v ...interface{}) {
	l.Logger.Println(v...)
	if l.CloudLogger != nil {
		l.CloudLogger.Error(v...)
	}
}

type CloudLogger struct {
	*logging.Logger
	*logging.Client
	mux *sync.Mutex
}

func CreateCloudLogger(ctx context.Context, projectID, loggerName string) (*CloudLogger, error) {
	logClient, err := logging.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &CloudLogger{
		Logger: logClient.Logger(loggerName),
		Client: logClient,
		mux:    &sync.Mutex{},
	}, nil
}

func (l *CloudLogger) Log(v ...interface{}) {
	l.mux.Lock()
	defer logger.mux.Unlock()

	l.Logger.Log(
		logging.Entry{Payload: fmt.Sprintln(v...)},
	)
}

func (l *CloudLogger) Logf(format string, v ...interface{}) {
	l.mux.Lock()
	defer logger.mux.Unlock()

	l.Logger.Log(
		logging.Entry{Payload: fmt.Sprintf(format, v...)},
	)
}

func (l *CloudLogger) Error(v ...interface{}) {
	l.mux.Lock()
	defer logger.mux.Unlock()

	l.Logger.Log(
		logging.Entry{
			Payload:  fmt.Sprintln(v...),
			Severity: logging.Error,
		},
	)
}

func (l *CloudLogger) Close() {
	l.Client.Close()
}
