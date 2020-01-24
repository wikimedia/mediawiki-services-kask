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
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Config represents an application-wide configuration.
type Config struct {
	ServiceName string `yaml:"service_name"`
	BaseURI     string `yaml:"base_uri"`
	Address     string `yaml:"listen_address"`
	Port        int    `yaml:"listen_port"`
	DefaultTTL  int    `yaml:"default_ttl"`
	LogLevel    string `yaml:"log_level"`
	OpenAPISpec string `yaml:"openapi_spec"`
	TLS         struct {
		CertPath string `yaml:"cert"`
		KeyPath  string `yaml:"key"`
	}

	Cassandra struct {
		Hosts          []string `yaml:"hosts"`
		Port           int      `yaml:"port"`
		Keyspace       string   `yaml:"keyspace"`
		Table          string   `yaml:"table"`
		LocalDC        string   `yaml:"local_dc"`
		QueryTimeout   int      `yaml:"query_timeout_ms"`
		ConnectTimeout int      `yaml:"connect_timeout_ms"`
		TLS            struct {
			CaPath   string `yaml:"ca"`
			CertPath string `yaml:"cert"`
			KeyPath  string `yaml:"key"`
		}
		Authentication struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		}
	}
}

// ReadConfig returns a new Config from a YAML file.
func ReadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewConfig(data)
}

// NewConfig returns a new Config from YAML serialized as bytes.
func NewConfig(data []byte) (*Config, error) {
	// Populate a new Config with sane defaults
	config := Config{
		ServiceName: "kask",
		BaseURI:     "/v1/",
		Address:     "localhost",
		Port:        8080,
		DefaultTTL:  86400,
		LogLevel:    "info",
	}
	config.Cassandra.Hosts = []string{"localhost"}
	config.Cassandra.Port = 9042
	config.Cassandra.Keyspace = "kask"
	config.Cassandra.Table = "values"
	config.Cassandra.QueryTimeout = 12000
	config.Cassandra.ConnectTimeout = 5000

	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return validate(&config)
}

func validate(config *Config) (*Config, error) {
	if !strings.HasSuffix(config.BaseURI, "/") {
		config.BaseURI += "/"
	}
	if !strings.HasPrefix(config.BaseURI, "/") {
		config.BaseURI = "/" + config.BaseURI
	}
	if config.DefaultTTL < 0 {
		return nil, errors.New("TTL must be a positive integer")
	}

	// Validate log level
	if err := validateLogLevel(config); err != nil {
		return nil, err
	}

	// Validate Kask TLS settings
	if err := validateKaskTLS(config); err != nil {
		return nil, err
	}

	// Validate Cassandra client authentication settings
	if err := validateCassandraAuthentication(config); err != nil {
		return nil, err
	}

	// Validate Cassandra client TLS settings
	if err := validateCassandraTLS(config); err != nil {
		return nil, err
	}

	// TODO: Consider some other validations
	return config, nil
}

// validateLogLevel ensures a valid log level
func validateLogLevel(config *Config) error {
	switch strings.ToUpper(config.LogLevel) {
	case "DEBUG", "INFO", "WARNING", "ERROR", "FATAL":
		return nil
	}
	return fmt.Errorf("Unsupported log level: %s", config.LogLevel)
}

// validateKaskTLS ensures a properly constructed TLS configuration.
func validateKaskTLS(config *Config) error {
	// Either CertPath and KeyPath are both zero (TLS not enabled), or both must be assigned.
	if !mutuallyInclusive(config.TLS.CertPath, config.TLS.KeyPath) {
		return errors.New("Kask cert/key values are mutually inclusive")
	}
	return nil
}

// validateCassandraAuthentication ensures a properly constructed Cassandra client authentication config.
func validateCassandraAuthentication(config *Config) error {
	auth := config.Cassandra.Authentication
	// Either username and password are both zero (authentication not enabled), or both must be assigned.
	if !mutuallyInclusive(auth.Username, auth.Password) {
		return errors.New("Cassandra username/password values are mutually inclusive")
	}
	return nil
}

// validateCassandraTLS ensures a properly constructed Cassandra client TLS configuration.
func validateCassandraTLS(config *Config) error {
	tls := config.Cassandra.TLS
	// If a ca is zero (unset), neither of cert/key can be.
	if tls.CaPath == "" && (tls.CertPath != "" || tls.KeyPath != "") {
		return errors.New("a Cassandra CA must be configured if key and cert are")
	}
	// If ca is set, then either both cert and key are, or neither are.
	if tls.CaPath != "" && !mutuallyInclusive(tls.CertPath, tls.KeyPath) {
		return errors.New("Cassandra TLS key/cert values are mutually inclusive")
	}
	return nil
}

// mutuallyInclusive returns true if its arguments are either both zero, or neither are.
func mutuallyInclusive(a string, b string) bool {
	if (a != "" && b == "") || (b != "" && a == "") {
		return false
	}
	return true
}
