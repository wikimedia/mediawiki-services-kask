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

	Cassandra struct {
		Hostname string `yaml:"hostname"`
		Port     int    `yaml:"port"`
		Keyspace string `yaml:"keyspace"`
		Table    string `yaml:"table"`
		TLS      struct {
			CaPath   string `yaml:"ca"`
			CertPath string `yaml:"cert"`
			KeyPath  string `yaml:"key"`
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
	}
	config.Cassandra.Hostname = "localhost"
	config.Cassandra.Port = 9042
	config.Cassandra.Keyspace = "kask"
	config.Cassandra.Table = "values"

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

	tls := config.Cassandra.TLS

	if tls.CaPath == "" && (tls.CertPath != "" || tls.KeyPath != "") {
		return nil, errors.New("a CA must be configured if key and cert are")
	}

	// TODO: Consider some other validations
	return config, nil
}
