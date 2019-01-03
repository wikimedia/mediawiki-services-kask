package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

const root string = "/sessions/v1/"
const service string = "kask"

var Log Logger

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

func getHandler(w http.ResponseWriter, r *http.Request, store Store, key string) {
	value, err := store.Get(key)
	if err != nil {
		// FIXME: This needs to differentiate between a failure to execute the SELECT, and
		// a record that is not found (which should result in 404, not 500).
		HttpError(w, InternelServerError(path.Join(root, key)))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value)
}

func postHandler(w http.ResponseWriter, r *http.Request, store Store, key string) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		HttpError(w, InternelServerError(path.Join(root, key)))
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err := store.Set(key, body); err != nil {
		Log.Error("Unable to persist value (%s)", err)
		HttpError(w, InternelServerError(path.Join(root, key)))
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
}

func putHandler(w http.ResponseWriter, r *http.Request, store Store, key string) {
	log.Println(key)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, store Store, key string) {
	if err := store.Delete(key); err != nil {
		HttpError(w, InternelServerError(path.Join(root, key)))
	}
}

// FIXME: Not good enough; Will violate Element of Least Surprise if
// base contains more than one path element (e.g. 'a/b/c/...').
func parseKey(url string) (string, error) {
	base := strings.Replace(url, root, "", 1)
	if base == "" {
		return base, errors.New("No key found")
	}
	return path.Base(url), nil
}

func dispatch(s Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := parseKey(r.URL.Path)
		if err != nil {
			HttpError(w, NotFound(path.Join(root, key)))
			return
		}

		switch r.Method {
		case http.MethodGet:
			getHandler(w, r, s, key)
		case http.MethodPost:
			postHandler(w, r, s, key)
		case http.MethodPut:
			putHandler(w, r, s, key)
		case http.MethodDelete:
			deleteHandler(w, r, s, key)
		default:
			HttpError(w, BadRequest(path.Join(root, key)))
		}
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

	Log = NewLog()
	Log.Info("Starting up...")

	http.HandleFunc(root, dispatch(store))
	log.Fatal(http.ListenAndServe(":8080", nil))

	defer store.Close()
}
