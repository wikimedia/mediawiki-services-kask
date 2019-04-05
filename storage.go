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
	"fmt"

	"github.com/gocql/gocql"
)

// Store is an interface to the underlying data store.
// Note: An interface exists to enable mocking in tests, not as a means of
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

func createSession(config *Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.Cassandra.Hostname)
	cluster.Port = config.Cassandra.Port
	cluster.Keyspace = config.Cassandra.Keyspace
	cluster.Consistency = gocql.LocalQuorum

	tlsConf := config.Cassandra.TLS

	if tlsConf.CaPath != "" {
		cluster.SslOpts = &gocql.SslOptions{
			CaPath: tlsConf.CaPath,
		}
		cluster.SslOpts.CertPath = tlsConf.CertPath
		cluster.SslOpts.KeyPath = tlsConf.KeyPath
	}

	authConf := config.Cassandra.Authentication

	if authConf.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: authConf.Username,
			Password: authConf.Password,
		}
	}

	return cluster.CreateSession()
}

// NewCassandraStore constructs new instances of CassandraStore.
func NewCassandraStore(config *Config) (*CassandraStore, error) {
	session, err := createSession(config)
	if err == nil {
		return &CassandraStore{session: session, Keyspace: config.Cassandra.Keyspace, Table: config.Cassandra.Table}, nil
	}
	return nil, err
}

// Set stores a new value associated with a key. Values expire after TTL
// seconds; Values with a TTL of 0 do not expire.
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
