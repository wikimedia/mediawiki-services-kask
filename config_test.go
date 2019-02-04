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
