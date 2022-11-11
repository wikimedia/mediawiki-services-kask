//go:build unit
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
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"text/template"

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

func setUp() (http.Handler, Store, error) {
	var store Store
	var config *Config
	var logger *Logger
	var err error

	store = newMockStore()
	if config, err = NewConfig([]byte("default_ttl: 300000")); err != nil {
		return nil, nil, err
	}
	if logger, err = NewLogger(os.Stdout, config.ServiceName, config.LogLevel); err != nil {
		return nil, nil, err
	}

	handler := &HTTPHandler{store, config, logger}
	return ValidatingKeyParserMiddleware(prefixURI, handler), store, nil
}

func setUpTesting(t *testing.T) (http.Handler, Store) {
	handler, store, err := setUp()
	if err != nil {
		t.Fatalf("Error encountered in test setup: %s", err)
		return nil, nil
	}
	return handler, store
}

func TestGetSuccess(t *testing.T) {
	handler, store := setUpTesting(t)

	req := httptest.NewRequest("GET", path.Join(prefixURI, "foo"), nil)
	res := httptest.NewRecorder()
	expected := "bar"

	store.Set("foo", []byte(expected), 300000)

	handler.ServeHTTP(res, req)

	AssertEquals(t, http.StatusOK, res.Code, "Incorrect status code")
	AssertEquals(t, expected, res.Body.String(), "Unexpected value")
}

func TestGetNotFound(t *testing.T) {
	handler, _ := setUpTesting(t)

	req := httptest.NewRequest("GET", path.Join(prefixURI, "cat"), nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	AssertEquals(t, http.StatusNotFound, res.Code, "Incorrect status code")
}

func TestPost(t *testing.T) {
	handler, store := setUpTesting(t)

	body := strings.NewReader("meow")
	req := httptest.NewRequest("POST", path.Join(prefixURI, "cat"), body)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	AssertEquals(t, http.StatusCreated, res.Code, "Incorrect status code")

	value, _ := store.Get("cat")
	expected := []byte("meow")

	if !bytes.Equal(value.Value, expected) {
		t.Errorf("POST added an unexpected value: got %v expected %v ", value, expected)
	}
}

func TestPostEmptyBody(t *testing.T) {
	handler, _ := setUpTesting(t)

	body := strings.NewReader("")
	req := httptest.NewRequest("POST", path.Join(prefixURI, "dog"), body)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	AssertEquals(t, http.StatusBadRequest, res.Code, "Incorrect status code")
}

func TestPut(t *testing.T) {
	handler, _ := setUpTesting(t)

	body := strings.NewReader("roar")
	req := httptest.NewRequest("PUT", path.Join(prefixURI, "cat"), body)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	AssertEquals(t, http.StatusBadRequest, res.Code, "Incorrect status code")
}

func TestDelete(t *testing.T) {
	handler, store := setUpTesting(t)

	req := httptest.NewRequest("DELETE", path.Join(prefixURI, "cat"), nil)
	res := httptest.NewRecorder()

	store.Set("cat", []byte("meow"), 300000)

	handler.ServeHTTP(res, req)

	AssertEquals(t, http.StatusNoContent, res.Code, "Incorrect status code")

	value, _ := store.Get("cat")

	if len(value.Value) > 0 {
		t.Errorf("DELETE did not remove key: cat and value: %s ", value.Value)
	}
}

func TestValidatingKeyParserMiddleware(t *testing.T) {
	testCases := []struct {
		url        string
		expected   string
		statusCode int
	}{
		{path.Join(prefixURI, "cat"), "cat", 200},
		{path.Join(prefixURI, "cat/dog"), "", 404},
		{prefixURI, "", 404},
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

			parser := ValidatingKeyParserMiddleware(prefixURI, handler)

			parser.ServeHTTP(rr, req)

			if rr.Code != tc.statusCode {
				AssertEquals(t, tc.statusCode, rr.Code, "Incorrect status code")
			}

		})
	}

}

func TestHealthz(t *testing.T) {
	handler := http.HandlerFunc(Healthz)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/healthz", nil))

	AssertEquals(t, http.StatusOK, rr.Code, "Incorrect status code")
}

func TestOpenAPI(t *testing.T) {
	config, err := NewConfig([]byte("openapi_spec: ./openapi.yaml"))
	if err != nil {
		t.Fatal("Unable to create Config instance")
	}

	logger, err := NewLogger(os.Stdout, config.ServiceName, config.LogLevel)
	if err != nil {
		t.Fatal("Unable to create Logger instance")
	}

	handler := http.HandlerFunc(OpenAPI(config, logger))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest("GET", "/openapi", nil))

	AssertEquals(t, http.StatusOK, rr.Code, "Incorrect status code")
	AssertEquals(t, "application/x-yaml", rr.Header().Get("Content-Type"), "Incorrect Content-Type header")

	// Calculate checksums for the response body and spec file, and validate that they match
	respHasher := sha256.New()
	fileHasher := sha256.New()

	if _, err := io.Copy(respHasher, rr.Body); err != nil {
		t.Fatalf("Error generating checksum of response: %s", err)
	}

	// Template the on-disk file (to match what the HTTP handler does)
	tmpl, err := template.New(path.Base(config.OpenAPISpec)).ParseFiles(config.OpenAPISpec)
	if err != nil {
		t.Fatalf("Unable to parse %s as template: %s", config.OpenAPISpec, err)
	}

	if err := tmpl.Execute(fileHasher, config); err != nil {
		t.Fatalf("Error generating checksum of OpenAPI spec: %s", err)
	}

	respSum := fmt.Sprintf("%x", respHasher.Sum(nil))
	fileSum := fmt.Sprintf("%x", fileHasher.Sum(nil))
	AssertEquals(t, fileSum, respSum, "OpenAPI response does not match file checksum")
}

func setUpBenchmark(t *testing.B) (http.Handler, Store) {
	handler, store, err := setUp()
	if err != nil {
		t.Fatalf("Error encountered in benchmark setup: %s", err)
		return nil, nil
	}
	return handler, store
}

func BenchmarkGet(b *testing.B) {
	handler, _ := setUpBenchmark(b)

	server := httptest.NewServer(handler)
	defer server.Close()

	url := fmt.Sprintf("%s/sessions/v1/%s", server.URL, RandString(16))
	val := strings.NewReader(RandString(32))
	client := &http.Client{}

	res, err := client.Post(url, "application/octet-stream", val)

	if err != nil {
		b.Fatalf("Error making POST request: %s", err)
		return
	}
	if res.StatusCode != http.StatusCreated {
		b.Fatalf("POST request failed (status: %d)", res.StatusCode)
		return
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res, err := client.Get(url)

		if err != nil {
			b.Fatalf("Client request failed: %s", err)
			break
		}
		if res.StatusCode != http.StatusOK {
			b.Fatalf("Request failed (status: %d)", res.StatusCode)
			break
		}
		if _, err = ioutil.ReadAll(res.Body); err != nil {
			b.Fatalf("Unable to read response body: %s", err)
			break
		}
		if err = res.Body.Close(); err != nil {
			b.Fatalf("Unable to close response body: %s", err)
			break
		}
	}
}

func BenchmarkPost(b *testing.B) {
	handler, _ := setUpBenchmark(b)

	server := httptest.NewServer(handler)
	defer server.Close()

	val := strings.NewReader(RandString(32))
	url := fmt.Sprintf("%s/sessions/v1/%s", server.URL, RandString(16))
	client := &http.Client{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res, err := client.Post(url, "application/octet-stream", val)

		if err != nil {
			b.Fatalf("Client request failed: %s", err)
			break
		}
		if res.StatusCode != http.StatusCreated {
			b.Fatalf("Request failed (status: %d)", res.StatusCode)
			break
		}

		// Reset the body reader
		val.Seek(0, io.SeekStart)
	}
}

func BenchmarkDelete(b *testing.B) {
	handler, _ := setUpBenchmark(b)

	server := httptest.NewServer(handler)
	defer server.Close()

	url := fmt.Sprintf("%s/sessions/v1/%s", server.URL, RandString(16))

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		b.Fatalf("Error creating request object: %s", err)
	}

	client := &http.Client{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res, err := client.Do(req)

		if err != nil {
			b.Fatalf("Client request failed: %s", err)
			break
		}
		if res.StatusCode != http.StatusNoContent {
			b.Fatalf("Request failed (status: %d)", res.StatusCode)
			break
		}
	}
}
