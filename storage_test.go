package main

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestSetGet(t *testing.T) {
	hostname := Getenv("CASSANDRA_HOST", "localhost")
	port := Getenv("CASSANDRA_PORT", "9042")
	keyspace := Getenv("CASSANDRA_KEYSPACE", "kask_test_keyspace")
	table := Getenv("CASSANDRA_TABLE", "test_table")

	portNum, err := strconv.Atoi(port)
	if err != nil {
		t.Errorf("%s is not a valid TCP port number!", port)
	}

	// Connect
	store, err := NewStore(hostname, portNum, keyspace, table)
	if err != nil {
		t.Errorf("Error connecting to data store (%s)", err)
	}

	key := randString(8)
	val := randString(32)

	// Write
	if err := store.Set(key, []byte(val)); err != nil {
		t.Errorf("Error storing value (%s)", err)
	}

	// Read
	if res, err := store.Get(key); err != nil {
		t.Errorf("Error retrieving value (%s)", err)
	} else {
		if string(res) != string(val) {
			t.Fail()
		}
	}
}
