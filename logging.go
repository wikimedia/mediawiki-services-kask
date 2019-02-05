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
	"log"
	"log/syslog"
)

// Log levels
const (
	LogDebug    = iota
	LogInfo     = iota
	LogWarning  = iota
	LogError    = iota
	LogCritical = iota
)

// Logger formats and delivers log messages.
type Logger struct {
	writer      *syslog.Writer
	serviceName string
}

// LogMessage represents JSON serializable log messages.
type LogMessage struct {
	Msg     string `json:"msg"`
	Appname string `json:"appname"`
}

// ceeString converts an interface into a JSON string prepended with @cee.
func ceeString(m interface{}) (string, error) {
	j, err := json.Marshal(m)

	return "@cee: " + string(j), err
}

// Log records a message at a specified level.
func (l *Logger) Log(i int, message LogMessage) {
	str, er := ceeString(message)

	// Handle the case where JSON serialization fails.
	if er != nil {
		err := l.writer.Err(fmt.Sprintf(`@cee: {"msg": "Error serializing log message: %v (%s)", "appname": "%s"}`, message, er, l.serviceName))
		if err != nil {
			log.Print(message)
		}
	}

	var err error

	switch i {
	case LogDebug:
		err = l.writer.Debug(str)
	case LogInfo:
		err = l.writer.Info(str)
	case LogWarning:
		err = l.writer.Warning(str)
	case LogError:
		err = l.writer.Err(str)
	case LogCritical:
		err = l.writer.Crit(str)
	default:
		l.Error("Invalid log level specified (%d); This is a bug!", i)
		err = l.writer.Err(str)
	}

	if err != nil {
		log.Print(message)
	}
}

// Critical logs messages of severity CRITICAL.
func (l *Logger) Critical(format string, v ...interface{}) {
	l.Log(LogCritical, LogMessage{fmt.Sprintf(format, v...), l.serviceName})
}

// Error logs messages of severity ERROR.
func (l *Logger) Error(format string, v ...interface{}) {
	l.Log(LogError, LogMessage{fmt.Sprintf(format, v...), l.serviceName})
}

// Warning logs messages of severity WARNING.
func (l *Logger) Warning(format string, v ...interface{}) {
	l.Log(LogWarning, LogMessage{fmt.Sprintf(format, v...), l.serviceName})
}

// Info logs messages of severity INFO.
func (l *Logger) Info(format string, v ...interface{}) {
	l.Log(LogInfo, LogMessage{fmt.Sprintf(format, v...), l.serviceName})
}

// Debug logs messages of severity DEBUG.
func (l *Logger) Debug(format string, v ...interface{}) {
	l.Log(LogDebug, LogMessage{fmt.Sprintf(format, v...), l.serviceName})
}

// NewLogger constructs new instances of Logger.
func NewLogger(serviceName string) *Logger {
	writer, err := syslog.Dial("", "", syslog.LOG_WARNING|syslog.LOG_DAEMON, serviceName)
	if err != nil {
		log.Fatal(err)
	}
	return &Logger{writer, serviceName}
}
