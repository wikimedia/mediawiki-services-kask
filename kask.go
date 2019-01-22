package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Close the database connection before returning from main()
	defer store.Close()

	keyMiddleware := NewParseKeyMiddleware(config.BaseURI)
	handler := HTTPHandler{store, config, logger}
	dispatcher := keyMiddleware(http.HandlerFunc(handler.Dispatch))
	address := fmt.Sprintf("%s:%d", config.Address, config.Port)

	logger.Info("Starting service as http://%s%s", address, config.BaseURI)

	http.Handle(config.BaseURI, dispatcher)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(address, nil))
}
