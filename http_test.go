package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func SetUp(t *testing.T) (http.Handler, Store) {
	store := newMockStore()
	config, _ := NewConfig([]byte("default_ttl: 300000"))
	logger := NewLogger("http_test")
	handler := HttpHandler{store, config, &logger}
	middleware := NewParseKeyMiddleware(prefixURI)
	handle := middleware(http.HandlerFunc(handler.Dispatch))

	return handle, store
}

func TestGetSuccess(t *testing.T) {
	handler, store := SetUp(t)
	url := fmt.Sprintf("%sfoo", prefixURI)
	req := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()
	expected := "bar"

	store.Set("foo", []byte("bar"), 300000)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code got %v expected %v", rr.Code, http.StatusOK)
	}

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v expected %v", rr.Body.String(), expected)
	}
}

func TestGetNotFound(t *testing.T) {
	handler, _ := SetUp(t)
	url := fmt.Sprintf("%scat", prefixURI)
	req := httptest.NewRequest("GET", url, nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("handler returned wrong status code got %v expected %v", rr.Code, http.StatusNotFound)
	}
}

func TestPost(t *testing.T) {
	handler, store := SetUp(t)
	url := fmt.Sprintf("%scat", prefixURI)
	body := strings.NewReader("meow")
	req := httptest.NewRequest("POST", url, body)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("handler returned wrong status code got %v expected %v", rr.Code, http.StatusCreated)
	}

	value, _ := store.Get("cat")
	expected := []byte("meow")

	if !bytes.Equal(value.Value, expected) {
		t.Errorf("POST added an unexpected value: got %v expected %v ", value, expected)
	}
}

func TestDelete(t *testing.T) {
	handler, store := SetUp(t)
	url := fmt.Sprintf("%scat", prefixURI)
	req := httptest.NewRequest("DELETE", url, nil)
	rr := httptest.NewRecorder()

	store.Set("cat", []byte("meow"), 300000)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("handler returned wrong status code got %v expected %v", rr.Code, http.StatusNoContent)
	}

	value, _ := store.Get("cat")

	if len(value.Value) > 0 {
		t.Errorf("DELETE did not remove key: cat and value: %s ", value.Value)
	}
}
