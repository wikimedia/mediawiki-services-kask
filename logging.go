package main

import (
	"fmt"
	"log"
	"log/syslog"
)

type Logger struct {
	writer *syslog.Writer
}

func (l *Logger) Critical(format string, v ...interface{}) {
	l.writer.Crit(fmt.Sprintf(format, v...))
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.writer.Err(fmt.Sprintf(format, v...))
}

func (l *Logger) Warning(format string, v ...interface{}) {
	l.writer.Warning(fmt.Sprintf(format, v...))
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.writer.Info(fmt.Sprintf(format, v...))
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.writer.Debug(fmt.Sprintf(format, v...))
}

func NewLogger(serviceName string) Logger {
	writer, err := syslog.Dial("", "", syslog.LOG_WARNING|syslog.LOG_DAEMON, serviceName)
	if err != nil {
		log.Fatal(err)
	}
	return Logger{writer}
}
