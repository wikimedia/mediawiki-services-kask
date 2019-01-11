package main

import (
	"fmt"

	"github.com/gocql/gocql"
)

type Store interface {
	Set(string, []byte, int) error
	Get(string) (Datum, error)
	Delete(string) error
	Close()
}

type CassandraStore struct {
	session  *gocql.Session
	Keyspace string
	Table    string
}

type Datum struct {
	Value []byte
	TTL   int
}

func createSession(hostname string, port int, keyspace string) (*gocql.Session, error) {
	cluster := gocql.NewCluster(hostname)
	cluster.Port = port
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.LocalQuorum
	return cluster.CreateSession()
}

func NewCassandraStore(hostname string, port int, keyspace string, table string) (*CassandraStore, error) {
	session, err := createSession(hostname, port, keyspace)
	if err == nil {
		return &CassandraStore{session: session, Keyspace: keyspace, Table: table}, nil
	}
	return nil, err
}

func (s *CassandraStore) Set(key string, value []byte, ttl int) error {
	query := fmt.Sprintf(`INSERT INTO "%s"."%s" (key, value) VALUES (?,?) USING TTL ?`, s.Keyspace, s.Table)
	return s.session.Query(query, key, value, ttl).Consistency(gocql.LocalQuorum).Exec()
}

func (s *CassandraStore) Get(key string) (Datum, error) {
	var value []byte
	var ttl int
	query := fmt.Sprintf(`SELECT value, TTL(value) as ttl FROM "%s"."%s" WHERE key = ?`, s.Keyspace, s.Table)
	err := s.session.Query(query, key).Consistency(gocql.LocalQuorum).Scan(&value, &ttl)
	return Datum{value, ttl}, err
}

func (s *CassandraStore) Delete(key string) error {
	query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE key = ?`, s.Keyspace, s.Table)
	return s.session.Query(query, key).Consistency(gocql.EachQuorum).Exec()
}

func (s *CassandraStore) Close() {
	s.session.Close()
}
