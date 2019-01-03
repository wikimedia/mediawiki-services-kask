package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

const root string = "/sessions/v1/"
const service string = "kask"

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
