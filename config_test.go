// +build unit

package main

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	var data = `
service_name:    kittens
base_uri:        /kittens/v1
listen_address:  172.17.0.2
listen_port:     8080
default_ttl:     1

cassandra:
  hostname: 172.17.0.3
  port: 9043
  keyspace: kittens
  table: data
`
	if config, err := NewConfig([]byte(data)); err == nil {
		AssertEquals(t, config.ServiceName, "kittens", "Service name")
		AssertEquals(t, config.BaseURI, "/kittens/v1/", "URI prefix")
		AssertEquals(t, config.Address, "172.17.0.2", "Bind address")
		AssertEquals(t, config.Port, 8080, "Port number")
		AssertEquals(t, config.Cassandra.Hostname, "172.17.0.3", "Cassandra hostname")
		AssertEquals(t, config.Cassandra.Port, 9043, "Cassandra port number")
		AssertEquals(t, config.Cassandra.Keyspace, "kittens", "Cassandra keyspace")
		AssertEquals(t, config.Cassandra.Table, "data", "Cassandra table name")
		AssertEquals(t, config.DefaultTTL, 1, "TTL value")
	} else {
		t.Errorf("Failed to read configuration data: %v", err)
	}

}

func TestNegativeTTL(t *testing.T) {
	if _, err := NewConfig([]byte("default_ttl: -1")); err == nil {
		t.Errorf("Negative TTLs are expected to fail validation!")
	}
}
