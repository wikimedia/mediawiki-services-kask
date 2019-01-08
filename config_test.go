package main

import (
	"reflect"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}, msg string) {
	if a == b {
		return
	}
	t.Errorf("%s: (%v (type %v) != %v (type %v))", msg, a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

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
		assertEqual(t, config.ServiceName, "kittens", "Unexpected service name")
		assertEqual(t, config.BaseURI, "/kittens/v1/", "Unexpected URI prefix")
		assertEqual(t, config.Address, "172.17.0.2", "Unexpected bind address")
		assertEqual(t, config.Port, 8080, "Unexpected port number")
		assertEqual(t, config.Cassandra.Hostname, "172.17.0.3", "Unexpected Cassandra host")
		assertEqual(t, config.Cassandra.Port, 9043, "Unexpected Cassandra port number")
		assertEqual(t, config.Cassandra.Keyspace, "kittens", "Unexpected Cassandra keyspace")
		assertEqual(t, config.Cassandra.Table, "data", "Unexpected Cassandra table name")
		assertEqual(t, config.DefaultTTL, 1, "Unexpected TTL value")
	} else {
		t.Errorf("Failed to read configuration data: %v", err)
	}

}

func TestNegativeTTL(t *testing.T) {
	if _, err := NewConfig([]byte("default_ttl: -1")); err == nil {
		t.Errorf("Negative TTLs are expected to fail validation!")
	}
}
