package main

import (
	"fmt"
	"github.com/gocql/gocql"
)

type Store struct {
	client   *gocql.Session
	Keyspace string
	Table    string
}

type Data struct {
	Key   string
	Value []byte
}

func CreateSession(hostname string, port int, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hostname)
	cluster.Port = port
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.LocalQuorum
	return cluster.CreateSession()
}

func NewStore(hostname string, port int, keyspace string, table string) (*Store, error) {
	if session, err := CreateSession(hostname, port, keyspace); err == nil {
		return &Store{client: session, Keyspace: keyspace, Table: table}, nil
	} else {
		return nil, err
	}
}

func (s *Store) Set(key string, value []byte) error {
	query := fmt.Sprintf(`INSERT INTO "%s"."%s" (key, value) VALUES (?,?)`, s.Keyspace, s.Table)
	return s.client.Query(query, key, value).Exec()
}

func (s *Store) Get(key string) ([]byte, error) {
	var value []byte
	query := fmt.Sprintf(`SELECT value FROM "%s"."%s" WHERE key = ?`, "kask_test_keyspace", "test_table")
	err := s.client.Query(query, key).Scan(&value)
	return value, err
}

func (s *Store) Close() {
	s.client.Close()
}
