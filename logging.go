/*
 * Copyright 2019 Clara Andrew-Wani <candrew@wikimedia.org>, Eric Evans <eevans@wikimedia.org>,
 * and Wikimedia Foundation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Log levels
const (
	LogDebug   = iota
	LogInfo    = iota
	LogWarning = iota
	LogError   = iota
	LogFatal   = iota
)

// Logger formats and delivers log messages.
type Logger struct {
	writer      io.Writer
	serviceName string
	logLevel    int
}

// LogMessage represents JSON serializable log messages.
type LogMessage struct {
	Msg       string `json:"msg"`
	Appname   string `json:"appname"`
	Time      string `json:"time"`
	Level     string `json:"level"`
	RequestID string `json:"request_id,omitempty"`
}

// ScopedLogger formats and delivers a Logger and optional LogMessage attributes.
type ScopedLogger struct {
	logger    *Logger
	requestID string
}

// Log creates a LogMessage at the specified level.
func (s *ScopedLogger) Log(level int, format string, v ...interface{}) {
	s.logger.log(level, func() LogMessage {
		return LogMessage{
			Msg:       fmt.Sprintf(format, v...),
			Time:      time.Now().Format(time.RFC3339),
			Appname:   s.logger.serviceName,
			Level:     LevelString(level),
			RequestID: s.requestID,
		}
	})
}

// RequestID records the request id and returns a ScopedLogger.
func (l *Logger) RequestID(id string) *ScopedLogger {
	return &ScopedLogger{logger: l, requestID: id}
}

// This is an internal implementation; The application should log messages
// using one of the level-specific methods, or a ScopedLogger as appropriate.
// Note: This method accepts a function that returns a LogMessage struct,
// instead of directly accepting a LogMessage, so that any costly string
// formatting can occur only if the message will be logged.
func (l *Logger) log(level int, msg func() LogMessage) {
	// Level must be one of the constants declared above; We do not allow ad hoc logging levels.
	if !validLevel(level) {
		l.Error("Invalid log level specified (%d); This is a bug!", level)
		level = LogError
	}

	// Skip if level is below what we're configured to log.
	if level < l.logLevel {
		return
	}

	message := msg()

	str, err := json.Marshal(message)

	// Handle the (unlikely) case where JSON serialization fails.
	if err != nil {
		l.write(fmt.Sprintf(`{"msg": "Error serializing log message: %v (%s)", "appname": "%s"}`, message, err, l.serviceName))
		return
	}

	// Log the messsage to the underlying io.Writer, one message per line.
	l.write(string(str))
}

// Fatal logs messages of severity FATAL.
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(LogFatal, l.basicLogMessage(LogFatal, format, v...))
}

// Error logs messages of severity ERROR.
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(LogError, l.basicLogMessage(LogError, format, v...))
}

// Warning logs messages of severity WARNING.
func (l *Logger) Warning(format string, v ...interface{}) {
	l.log(LogWarning, l.basicLogMessage(LogWarning, format, v...))
}

// Info logs messages of severity INFO.
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(LogInfo, l.basicLogMessage(LogInfo, format, v...))
}

// Debug logs messages of severity DEBUG.
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(LogDebug, l.basicLogMessage(LogDebug, format, v...))
}

func (l *Logger) write(s string) {
	// TODO: Should error handling be added to this? Our io.Writer will likely always be
	// os.Stdout, what would we do if unable to write to stdout?
	fmt.Fprintln(l.writer, s)
}

// This is an (internal) utility method for creating simple LogMessage (functions).
func (l *Logger) basicLogMessage(level int, format string, v ...interface{}) func() LogMessage {
	return func() LogMessage {
		return LogMessage{
			Msg:     fmt.Sprintf(format, v...),
			Time:    time.Now().Format(time.RFC3339),
			Appname: l.serviceName,
			Level:   LevelString(level),
		}
	}
}

func validLevel(level int) bool {
	switch level {
	case LogDebug, LogInfo, LogWarning, LogError, LogFatal:
		return true
	}
	return false
}

// LevelString converts log integers to strings
func LevelString(level int) string {
	switch level {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarning:
		return "WARNING"
	case LogError:
		return "ERROR"
	case LogFatal:
		return "FATAL"
	default:
		return ""
	}
}

// NewLogger creates a new instance of Logger
func NewLogger(writer io.Writer, serviceName string, logLevel string) (*Logger, error) {
	var level int

	switch strings.ToUpper(logLevel) {
	case LevelString(LogDebug):
		level = LogDebug
	case LevelString(LogInfo):
		level = LogInfo
	case LevelString(LogWarning):
		level = LogWarning
	case LevelString(LogError):
		level = LogError
	case LevelString(LogFatal):
		level = LogFatal
	default:
		return nil, fmt.Errorf("Unsupported log level: %s", logLevel)
	}

	return &Logger{writer, serviceName, level}, nil
}
