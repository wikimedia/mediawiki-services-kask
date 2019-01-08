package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var confFile = flag.String("config", "/etc/kask/config.yaml", "Path to the configuration file")

func main() {
	flag.Parse()

	config, err := ReadConfig(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	logger := NewLogger(config.ServiceName)

	store, err := NewCassandraStore(config.Cassandra.Hostname, config.Cassandra.Port, config.Cassandra.Keyspace, config.Cassandra.Table)
	if err != nil {
		logger.Error("Error connecting to Cassandra: %s", err)
		log.Fatal("Error connecting to Cassandra: ", err)
	}

	handler := HttpHandler{store, &logger}
	address := fmt.Sprintf("%s:%d", config.Address, config.Port)

	logger.Info("Starting service as http://%s%s", address, config.BaseUri)

	http.Handle(config.BaseUri, ParseKeyMiddleware(config.BaseUri, http.HandlerFunc(handler.Dispatch)))
	log.Fatal(http.ListenAndServe(address, nil))

	defer store.Close()
}
