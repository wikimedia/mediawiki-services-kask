package main

import (
	"fmt"

	"github.com/gocql/gocql"
)

type Store struct {
	session  *gocql.Session
	Keyspace string
	Table    string
}

func createSession(hostname string, port int, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hostname)
	cluster.Port = port
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.LocalQuorum
	return cluster.CreateSession()
}

func NewStore(hostname string, port int, keyspace string, table string) (*Store, error) {
	if session, err := createSession(hostname, port, keyspace); err == nil {
		return &Store{session: session, Keyspace: keyspace, Table: table}, nil
	} else {
		return nil, err
	}
}

func (s *Store) Set(key string, value []byte) error {
	query := fmt.Sprintf(`INSERT INTO "%s"."%s" (key, value) VALUES (?,?)`, s.Keyspace, s.Table)
	return s.session.Query(query, key, value).Exec()
}

func (s *Store) Get(key string) ([]byte, error) {
	var value []byte
	query := fmt.Sprintf(`SELECT value FROM "%s"."%s" WHERE key = ?`, s.Keyspace, s.Table)
	err := s.session.Query(query, key).Scan(&value)
	return value, err
}

func (s *Store) Delete(key string) error {
	query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE key = ?`, s.Keyspace, s.Table)
	return s.session.Query(query, key).Exec()
}

func (s *Store) Close() {
	s.session.Close()
}
