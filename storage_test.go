// +build functional

package main

import (
	"math/rand"
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

func TestSetGetDelete(t *testing.T) {
	config, err := ReadConfig(*confFile)
	if err != nil {
		t.Error(err)
	}

	// Connect
	store, err := NewCassandraStore(config.Cassandra.Hostname, config.Cassandra.Port, config.Cassandra.Keyspace, config.Cassandra.Table)
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

	// Delete
	if err := store.Delete(key); err != nil {
		t.Errorf("Error deleting value (%s)", err)
	}

	// Read
	if _, err := store.Get(key); err == nil {
		t.Fail()
	}
}
