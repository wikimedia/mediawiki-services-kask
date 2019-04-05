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
	"fmt"
	"testing"
)

func TestNewConfig(t *testing.T) {
	var data = `
service_name:    kittens
base_uri:        /kittens/v1
listen_address:  172.17.0.2
listen_port:     8888
default_ttl:     1
log_level:       error

tls:
  cert: /path/to/cert
  key:  /path/to/key

cassandra:
  hostname: 172.17.0.3
  port: 9043
  keyspace: kittens
  table: data
  authentication:
    username: myuser
    password: mypass
  tls:
    ca:   /path/to/ca
    key:  /path/to/key
    cert: /path/to/cert
`
	if config, err := NewConfig([]byte(data)); err == nil {
		AssertEquals(t, config.ServiceName, "kittens", "Service name")
		AssertEquals(t, config.BaseURI, "/kittens/v1/", "URI prefix")
		AssertEquals(t, config.Address, "172.17.0.2", "Bind address")
		AssertEquals(t, config.Port, 8888, "Port number")
		AssertEquals(t, config.TLS.CertPath, "/path/to/cert", "Kask TLS cert path name")
		AssertEquals(t, config.TLS.KeyPath, "/path/to/key", "Kask TLS key path name")
		AssertEquals(t, config.DefaultTTL, 1, "TTL value")
		AssertEquals(t, config.LogLevel, "error", "Log level")
		AssertEquals(t, config.Cassandra.Hostname, "172.17.0.3", "Cassandra hostname")
		AssertEquals(t, config.Cassandra.Port, 9043, "Cassandra port number")
		AssertEquals(t, config.Cassandra.Keyspace, "kittens", "Cassandra keyspace")
		AssertEquals(t, config.Cassandra.Table, "data", "Cassandra table name")
		AssertEquals(t, config.Cassandra.Authentication.Username, "myuser", "Cassandra username")
		AssertEquals(t, config.Cassandra.Authentication.Password, "mypass", "Cassandra password")
		AssertEquals(t, config.Cassandra.TLS.CaPath, "/path/to/ca", "Cassandra TLS CA path name")
		AssertEquals(t, config.Cassandra.TLS.KeyPath, "/path/to/key", "Cassandra TLS key path name")
		AssertEquals(t, config.Cassandra.TLS.CertPath, "/path/to/cert", "Cassandra TLS cert path name")
	} else {
		t.Errorf("Failed to read configuration data: %v", err)
	}

}

func TestNegativeTTL(t *testing.T) {
	if _, err := NewConfig([]byte("default_ttl: -1")); err == nil {
		t.Errorf("Negative TTLs are expected to fail validation!")
	}
}

func TestInvalidLogLevel(t *testing.T) {
	if _, err := NewConfig([]byte("log_level: emergency")); err == nil {
		t.Errorf("Invalid/unsupported log levels are expected to fail validation!")
	}
}

func TestKaskTLSValidation(t *testing.T) {
	t.Run("Unset cert w/ assigned key", func(t *testing.T) {
		var data = []byte(fmt.Sprintf("tls:\n    key: /path/to/key"))
		if _, err := NewConfig(data); err == nil {
			t.Errorf("Unset cert with assigned key expected to fail validation!")
		}
	})

	t.Run("Unset key w/ assigned cert", func(t *testing.T) {
		var data = []byte(fmt.Sprintf("tls:\n    cert: /path/to/cert"))
		if _, err := NewConfig(data); err == nil {
			t.Errorf("Unset key with assigned cert expected to fail validation!")
		}
	})
}

func TestCassandraAuthenticationValidation(t *testing.T) {
	var data = `
cassandra:
  authentication:
    %s: xxxxxxx
`
	t.Run("Username w/o password", func(t *testing.T) {
		if _, err := NewConfig([]byte(fmt.Sprint(data, "username"))); err == nil {
			t.Errorf("Unset password and assigned username expected to fail validation!")
		}

	})

	t.Run("Password w/o username", func(t *testing.T) {
		if _, err := NewConfig([]byte(fmt.Sprint(data, "password"))); err == nil {
			t.Errorf("Unset username and assigned password expected to fail validation!")
		}

	})
}

func TestCaValidation(t *testing.T) {
	t.Run("Unset CA w/ assigned key", func(t *testing.T) {
		var data = []byte(fmt.Sprintf("cassandra:\n  tls:\n    key: /path/to/key"))
		if _, err := NewConfig(data); err == nil {
			t.Errorf("Unset CA with assigned key expected to fail validation!")
		}
	})

	t.Run("Unset CA w/ assigned cert", func(t *testing.T) {
		var data = []byte(fmt.Sprintf("cassandra:\n  tls:\n    cert: /path/to/cert"))
		if _, err := NewConfig(data); err == nil {
			t.Errorf("Unset CA with assigned cert expected to fail validation!")
		}
	})
}
