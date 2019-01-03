package main

import (
	"github.com/gocql/gocql"
)

type mockStore struct {
	data map[string][]byte
}

func (m *mockStore) Set(key string, value []byte) error {
	m.data[key] = value
	return nil
}

func (m *mockStore) Get(key string) ([]byte, error) {
	if value, ok := m.data[key]; ok {
		return m.data[key], nil
	} else {
		return value, gocql.ErrNotFound
	}
}

func (m *mockStore) Delete(key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockStore) Close(key string) {
	return
}

func NewMockStore() *mockStore {
	return &mockStore{make(map[string][]byte)}
}
