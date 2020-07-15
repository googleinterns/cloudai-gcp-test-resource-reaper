// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/logging"
)

// Logger handles writing local logs to a file and cloud logs to Stackdriver.
type Logger struct {
	*CloudLogger
	*log.Logger
}

var logger *Logger

// CreateLogger initializes the logger for the server. The logs will be written to a local
// file called logs.txt.
func CreateLogger() error {
	logFile, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	fileLogger := log.New(logFile, "", log.Ldate|log.Ltime)
	logger = &Logger{Logger: fileLogger}
	return nil
}

// Log outputs to the necessary logs. Arguments are handled in the manner of fmt.Println.
func Log(v ...interface{}) {
	logger.log(v...)
}

// Logf takes a format string and message and writes it to the necessary logs. Arguments are
// handled in the manner of fmt.Printf.
func Logf(format string, v ...interface{}) {
	logger.logf(format, v...)
}

// Error outputs an error to the necessary logs.
func Error(v ...interface{}) {
	logger.error(v...)
}

// Close closes the logger.
func Close() {
	if logger.CloudLogger != nil {
		logger.CloudLogger.closeLogger()
	}
}

// AddCloudLogger adds stackdriver logging to the logger in the given project and log name.
func AddCloudLogger(ctx context.Context, projectID, loggerName string) error {
	cloudLogger, err := createCloudLogger(ctx, projectID, loggerName)
	if err != nil {
		return err
	}
	logger.CloudLogger = cloudLogger
	return nil
}

func (l *Logger) log(v ...interface{}) {
	l.Logger.Println(v...)
	if l.CloudLogger != nil {
		l.CloudLogger.log(v...)
	}
}

func (l *Logger) logf(format string, v ...interface{}) {
	l.Logger.Printf(format, v...)
	if l.CloudLogger != nil {
		l.CloudLogger.logf(format, v...)
	}
}

func (l *Logger) error(v ...interface{}) {
	l.Logger.Println(v...)
	if l.CloudLogger != nil {
		l.CloudLogger.error(v...)
	}
}

// CloudLogger handles writing logs to stackdriver.
type CloudLogger struct {
	*logging.Logger
	*logging.Client
	mux *sync.Mutex
}

func createCloudLogger(ctx context.Context, projectID, loggerName string) (*CloudLogger, error) {
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

func (l *CloudLogger) log(v ...interface{}) {
	l.mux.Lock()
	defer logger.mux.Unlock()

	l.Logger.Log(
		logging.Entry{Payload: fmt.Sprintln(v...)},
	)
}

func (l *CloudLogger) logf(format string, v ...interface{}) {
	l.mux.Lock()
	defer logger.mux.Unlock()

	l.Logger.Log(
		logging.Entry{Payload: fmt.Sprintf(format, v...)},
	)
}

func (l *CloudLogger) error(v ...interface{}) {
	l.mux.Lock()
	defer logger.mux.Unlock()

	l.Logger.Log(
		logging.Entry{
			Payload:  fmt.Sprintln(v...),
			Severity: logging.Error,
		},
	)
}

func (l *CloudLogger) closeLogger() {
	l.Client.Close()
}
