//go:build functional
// +build functional

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
	"testing"
	"time"

	"github.com/gocql/gocql"
)

const defaultTTL = 300

func setup(t *testing.T) (*CassandraStore, error) {
	config, err := ReadConfig(*confFile)
	if err != nil {
		return nil, err
	}

	// Connect
	store, err := NewCassandraStore(config)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func TestSetGetDelete(t *testing.T) {
	store, err := setup(t)
	if err != nil {
		t.Errorf("Test setup failure: %s", err)
		return
	}

	key := RandString(8)
	val := RandString(32)

	t.Run("SET", func(t *testing.T) {
		if err := store.Set(key, []byte(val), defaultTTL); err != nil {
			t.Errorf("Error storing value (%s)", err)
		}
	})

	t.Run("GET#01", func(t *testing.T) {
		if res, err := store.Get(key); err != nil {
			t.Errorf("Error retrieving value (%s)", err)
		} else {
			if string(res.Value) != string(val) {
				t.Fail()
			}
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		if err := store.Delete(key); err != nil {
			t.Errorf("Error deleting value (%s)", err)
		}
	})

	t.Run("GET#02", func(t *testing.T) {
		if _, err := store.Get(key); err == nil {
			t.Fail()
		}
	})
}

func TestTTL(t *testing.T) {
	store, err := setup(t)
	if err != nil {
		t.Errorf("Test setup failure: %s", err)
		return
	}

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
