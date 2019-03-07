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
	"time"
)

// Log levels
const (
	LogDebug   = "DEBUG"
	LogInfo    = "INFO"
	LogWarning = "WARN"
	LogError   = "ERROR"
	LogFatal   = "FATAL"
)

// Logger formats and delivers log messages.
type Logger struct {
	writer      io.Writer
	serviceName string
}

// LogMessage represents JSON serializable log messages.
type LogMessage struct {
	Msg     string `json:"msg"`
	Appname string `json:"appname"`
	Time    string `json:"time"`
	Level   string `json:"level"`
	ReqID   string `json:"request_id,omitempty"`
}

// ScopedLogger formats and delivers a Logger and optional LogMessage attributes.
type ScopedLogger struct {
	logger    *Logger
	requestID string
}

// Log creates a LogMessage at the specified level.
func (s *ScopedLogger) Log(level string, format string, v ...interface{}) {
	message := LogMessage{Msg: fmt.Sprintf(format, v...), ReqID: s.requestID}
	s.logger.log(level, message)
}

// RequestID records the request id and returns a ScopedLogger.
func (l *Logger) RequestID(id string) *ScopedLogger {
	return &ScopedLogger{logger: l, requestID: id}
}

// Log populates the remaining attributes of LogMessage at a specified level and logs the message.
func (l *Logger) log(level string, message LogMessage) {
	// Level must be one of the constants declared above; We do not allow ad hoc logging levels.
	if !validLevel(level) {
		l.Error("Invalid log level specified (%s); This is a bug!", level)
		level = LogError
	}

	// RFC3339 reads like a stricter version of ISO8601
	message.Time = time.Now().Format(time.RFC3339)
	message.Appname = l.serviceName
	message.Level = level

	str, err := json.Marshal(message)

	// Handle the (unlikely) case where JSON serialization fails.
	if err != nil {
		l.write(fmt.Sprintf(`{"msg": "Error serializing log message: %v (%s)", "appname": "%s"}`, message, err, l.serviceName))
		return
	}

	// Log the messsage to the underlying io.Writer, one message per line.
	l.write(string(str))
}

// Fatal logs messages of severity CRITICAL.
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(LogFatal, LogMessage{Msg: fmt.Sprintf(format, v...)})
}

// Error logs messages of severity ERROR.
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(LogError, LogMessage{Msg: fmt.Sprintf(format, v...)})
}

// Warning logs messages of severity WARNING.
func (l *Logger) Warning(format string, v ...interface{}) {
	l.log(LogWarning, LogMessage{Msg: fmt.Sprintf(format, v...)})
}

// Info logs messages of severity INFO.
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(LogInfo, LogMessage{Msg: fmt.Sprintf(format, v...)})
}

// Debug logs messages of severity DEBUG.
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(LogDebug, LogMessage{Msg: fmt.Sprintf(format, v...)})
}

func (l *Logger) write(s string) {
	// TODO: Should error handling be added to this? Our io.Writer will likely always be
	// os.Stdout, what would we do if unable to write to stdout?
	fmt.Fprintln(l.writer, s)
}

func validLevel(level string) bool {
	switch level {
	case LogDebug, LogInfo, LogWarning, LogError, LogFatal:
		return true
	}
	return false
}
