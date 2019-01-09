// +build functional

package main

import (
	"testing"
	"time"

	"github.com/gocql/gocql"
)

const defaultTTL = 300

func setup(t *testing.T) *CassandraStore {
	config, err := ReadConfig(*confFile)
	if err != nil {
		t.Error(err)
	}

	// Connect
	store, err := NewCassandraStore(config.Cassandra.Hostname, config.Cassandra.Port, config.Cassandra.Keyspace, config.Cassandra.Table)
	if err != nil {
		t.Errorf("Error connecting to data store (%s)", err)
	}

	return store
}

func TestSetGetDelete(t *testing.T) {
	store := setup(t)

	key := RandString(8)
	val := RandString(32)

	// Write
	if err := store.Set(key, []byte(val), defaultTTL); err != nil {
		t.Errorf("Error storing value (%s)", err)
	}

	// Read
	if res, err := store.Get(key); err != nil {
		t.Errorf("Error retrieving value (%s)", err)
	} else {
		if string(res.Value) != string(val) {
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

func TestTTL(t *testing.T) {
	store := setup(t)

	key := RandString(8)
	val := RandString(32)

	// Write a value with TTL of 5 seconds
	if err := store.Set(key, []byte(val), 5); err != nil {
		t.Errorf("Error storing value (%s)", err)
	}

	// Read
	if res, err := store.Get(key); err != nil {
		t.Errorf("Error retrieving value (%s)", err)
	} else {
		if string(res.Value) != string(val) {
			t.Fail()
		}
		if res.TTL > 300 || res.TTL < 0 {
			t.Fail()
		}
	}

	time.Sleep(5001 * time.Millisecond)

	// Read again after (at least) 5 seconds and 1 millisecond
	if res, err := store.Get(key); err != gocql.ErrNotFound {
		t.Errorf("Expected value to have expired but result (%v) returned", res)
	}
}
