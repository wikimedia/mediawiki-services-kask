package main

import (
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"strconv"
)

const root string = "/sessions/v1/"
const service string = "kask"

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

func NewLog() Logger {
	writer, err := syslog.Dial("", "", syslog.LOG_WARNING|syslog.LOG_DAEMON, service)
	if err != nil {
		log.Fatal(err)
	}
	return Logger{writer}
}

func Getenv(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	} else {
		return value
	}
}

func main() {
	hostname := Getenv("CASSANDRA_HOST", "localhost")
	port := Getenv("CASSANDRA_PORT", "9042")
	keyspace := Getenv("CASSANDRA_KEYSPACE", "kask_test_keyspace")
	table := Getenv("CASSANDRA_TABLE", "test_table")

	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("%s is not a valid TCP port number!", port)
	}

	// TODO: Handle errors...
	store, _ := NewCassandraStore(hostname, portNum, keyspace, table)

	logger := NewLog()
	logger.Info("Starting up...")

	handler := HttpHandler{store, &logger}

	http.Handle(root, ParseKeyMiddleware(root, http.HandlerFunc(handler.Dispatch)))
	log.Fatal(http.ListenAndServe(":8080", nil))

	defer store.Close()
}
