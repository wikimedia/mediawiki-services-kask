package main

import (
	"fmt"
	"log"
	"log/syslog"
)

// Logger formats and delivers log messages.
type Logger struct {
	writer *syslog.Writer
}

// Critical logs messages of severity CRITICAL.
func (l *Logger) Critical(format string, v ...interface{}) {
	l.writer.Crit(fmt.Sprintf(format, v...))
}

// Error logs messages of severity ERROR.
func (l *Logger) Error(format string, v ...interface{}) {
	l.writer.Err(fmt.Sprintf(format, v...))
}

// Warning logs messages of severity WARNING.
func (l *Logger) Warning(format string, v ...interface{}) {
	l.writer.Warning(fmt.Sprintf(format, v...))
}

// Info logs messages of severity INFO.
func (l *Logger) Info(format string, v ...interface{}) {
	l.writer.Info(fmt.Sprintf(format, v...))
}

// Debug logs messages of severity DEBUG.
func (l *Logger) Debug(format string, v ...interface{}) {
	l.writer.Debug(fmt.Sprintf(format, v...))
}

// NewLogger constructs new instances of Logger.
func NewLogger(serviceName string) *Logger {
	writer, err := syslog.Dial("", "", syslog.LOG_WARNING|syslog.LOG_DAEMON, serviceName)
	if err != nil {
		log.Fatal(err)
	}
	return &Logger{writer}
}
