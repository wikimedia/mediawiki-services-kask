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
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	confFile = flag.String("config", "/etc/kask/config.yaml", "Path to the configuration file")

	httpReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method"},
	)

	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of latencies for requests, partitioned by status code and HTTP method.",
			Buckets: []float64{.001, .0025, .0050, .01, .025, .050, .10, .25, .50, 1},
		},
		[]string{"code", "method"},
	)
)

func init() {
	prometheus.MustRegister(httpReqs, duration)
}

func main() {
	flag.Parse()

	config, err := ReadConfig(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	logger := NewLogger(config.ServiceName)

	store, err := NewCassandraStore(config)
	if err != nil {
		logger.Error("Error connecting to Cassandra: %s", err)
		log.Fatal("Error connecting to Cassandra: ", err)
	}

	// Close the database connection before returning from main()
	defer store.Close()

	// Kask CRUD operations
	handler := &HTTPHandler{store, config, logger}

	// Wrap in middlewares
	dispatcher := NewParseKeyMiddleware(config.BaseURI)(handler)
	dispatcher = promhttp.InstrumentHandlerCounter(httpReqs, dispatcher)
	dispatcher = promhttp.InstrumentHandlerDuration(duration, dispatcher)

	listen := fmt.Sprintf("%s:%d", config.Address, config.Port)
	logger.Info("Starting service as http://%s%s", listen, config.BaseURI)

	http.Handle(config.BaseURI, dispatcher)
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(listen, nil))
}
