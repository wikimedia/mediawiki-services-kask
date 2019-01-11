package main

import (
	"fmt"

	"github.com/gocql/gocql"
)

// Store is an interface to the underlying data store.
// Note: For the most part, an interface exists only to enable mocking in tests, not as a means of
// making storage pluggable.
type Store interface {
	Set(string, []byte, int) error
	Get(string) (Datum, error)
	Delete(string) error
	Close()
}

// CassandraStore provides access to storage using Apache Cassandra.
type CassandraStore struct {
	session  *gocql.Session
	Keyspace string
	Table    string
}

// Datum represents a value returned from storage.
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

// NewCassandraStore constructs new instances of CassandraStore.
func NewCassandraStore(hostname string, port int, keyspace string, table string) (*CassandraStore, error) {
	session, err := createSession(hostname, port, keyspace)
	if err == nil {
		return &CassandraStore{session: session, Keyspace: keyspace, Table: table}, nil
	}
	return nil, err
}

// Set stores a new value associated with a key.
// Values expire after TTL seconds; Values with a TTL of 0 do not expire.
func (s *CassandraStore) Set(key string, value []byte, ttl int) error {
	query := fmt.Sprintf(`INSERT INTO "%s"."%s" (key, value) VALUES (?,?) USING TTL ?`, s.Keyspace, s.Table)
	return s.session.Query(query, key, value, ttl).Consistency(gocql.LocalQuorum).Exec()
}

// Get retrieves a value associated with a key.
func (s *CassandraStore) Get(key string) (Datum, error) {
	var value []byte
	var ttl int
	query := fmt.Sprintf(`SELECT value, TTL(value) as ttl FROM "%s"."%s" WHERE key = ?`, s.Keyspace, s.Table)
	err := s.session.Query(query, key).Consistency(gocql.LocalQuorum).Scan(&value, &ttl)
	return Datum{value, ttl}, err
}

// Delete removes a value associated with a key.
func (s *CassandraStore) Delete(key string) error {
	query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE key = ?`, s.Keyspace, s.Table)
	return s.session.Query(query, key).Consistency(gocql.EachQuorum).Exec()
}

// Close terminates the underlying session to Cassandra (disconnects).
func (s *CassandraStore) Close() {
	s.session.Close()
}
