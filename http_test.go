// +build unit

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
	"bytes"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/gocql/gocql"
)

type mockStore struct {
	data map[string][]byte
}

func (m *mockStore) Set(key string, value []byte, ttl int) error {
	m.data[key] = value
	return nil
}

func (m *mockStore) Get(key string) (Datum, error) {
	if value, ok := m.data[key]; ok {
		return Datum{value, 0}, nil
	}
	return Datum{nil, 0}, gocql.ErrNotFound
}

func (m *mockStore) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockStore) Close() {
	return
}

func newMockStore() *mockStore {
	return &mockStore{make(map[string][]byte)}
}

const prefixURI = "/sessions/v1/"

func setUp(t *testing.T) (http.Handler, Store) {
	store := newMockStore()
	config, _ := NewConfig([]byte("default_ttl: 300000"))
	logger := NewLogger("http_test")
	handler := NewParseKeyMiddleware(prefixURI)(&HTTPHandler{store, config, logger})

	return handler, store
}

func TestGetSuccess(t *testing.T) {
	handler, store := setUp(t)
	url := path.Join(prefixURI, "foo")
	req := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	expected := "bar"

	store.Set("foo", []byte("bar"), 300000)

	handler.ServeHTTP(rr, req)

	AssertEquals(t, http.StatusOK, rr.Code, "Incorrect status code")
	AssertEquals(t, expected, rr.Body.String(), "Unexpected value")
}

func TestGetNotFound(t *testing.T) {
	handler, _ := setUp(t)
	url := path.Join(prefixURI, "cat")
	req := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	AssertEquals(t, http.StatusNotFound, rr.Code, "Incorrect status code")
}

func TestPost(t *testing.T) {
	handler, store := setUp(t)
	url := path.Join(prefixURI, "cat")
	body := strings.NewReader("meow")
	req := httptest.NewRequest("POST", url, body)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	AssertEquals(t, http.StatusCreated, rr.Code, "Incorrect status code")

	value, _ := store.Get("cat")
	expected := []byte("meow")

	if !bytes.Equal(value.Value, expected) {
		t.Errorf("POST added an unexpected value: got %v expected %v ", value, expected)
	}
}

func TestPut(t *testing.T) {
	handler, _ := setUp(t)
	url := path.Join(prefixURI, "cat")
	body := strings.NewReader("roar")
	req := httptest.NewRequest("PUT", url, body)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	AssertEquals(t, http.StatusBadRequest, rr.Code, "Incorrect status code")
}

func TestDelete(t *testing.T) {
	handler, store := setUp(t)
	url := path.Join(prefixURI, "cat")
	req := httptest.NewRequest("DELETE", url, nil)
	rr := httptest.NewRecorder()

	store.Set("cat", []byte("meow"), 300000)

	handler.ServeHTTP(rr, req)

	AssertEquals(t, http.StatusNoContent, rr.Code, "Incorrect status code")

	value, _ := store.Get("cat")

	if len(value.Value) > 0 {
		t.Errorf("DELETE did not remove key: cat and value: %s ", value.Value)
	}
}

func TestNewParseKeyMiddleware(t *testing.T) {
	testCases := []struct {
		url        string
		expected   string
		statusCode int
	}{
		{path.Join(prefixURI, "cat"), "cat", 200},
		{path.Join(prefixURI, "cat/dog"), "", 404},
		{"/", "", 404},
		{"/something/else", "", 404},
		{path.Join(prefixURI, "foo%3Fbar"), "foo?bar", 200},
		{path.Join(prefixURI, "foo?bar"), "", 400},
		{path.Join(prefixURI, "cat%20dog"), "cat dog", 200},
	}
	for _, tc := range testCases {
		t.Run(tc.url, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				key := r.Context().Value(kaskKey).(string)
				AssertEquals(t, tc.expected, key, "Incorrect key parsed")
			})

			req := httptest.NewRequest("GET", tc.url, nil)
			rr := httptest.NewRecorder()

			parser := NewParseKeyMiddleware(prefixURI)(handler)

			parser.ServeHTTP(rr, req)

			if rr.Code != tc.statusCode {
				AssertEquals(t, tc.statusCode, rr.Code, "Incorrect status code")
			}

		})
	}

}
