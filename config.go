package main

import (
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ServiceName string `yaml:"service_name"`
	BaseUri     string `yaml:"base_uri"`
	Address     string `yaml:"listen_address"`
	Port        int    `yaml:"listen_port"`

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
		BaseUri:     "/v1/",
		Address:     "localhost",
		Port:        8080,
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
	if !strings.HasSuffix(config.BaseUri, "/") {
		config.BaseUri += "/"
	}
	if !strings.HasPrefix(config.BaseUri, "/") {
		config.BaseUri = "/" + config.BaseUri
	}
	// TODO: Consider other validations
	return config, nil
}
