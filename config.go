package main

import (
	"errors"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

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
	}
}

func ReadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return NewConfig(data)
}

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
	// TODO: Consider some other validations
	return config, nil
}
