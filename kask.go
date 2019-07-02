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
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	confFile = flag.String("config", "/etc/kask/config.yaml", "Path to the configuration file")

	promHTTPReqsCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method"},
	)

	promDurationHistoVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of latencies for requests, partitioned by status code and HTTP method.",
			Buckets: []float64{.001, .0025, .0050, .01, .025, .050, .10, .25, .50, 1},
		},
		[]string{"code", "method"},
	)

	// These values are passed in at build time using -ldflags
	version   = "unknown"
	gitTag    = "unknown"
	buildHost = "unknown"
	buildDate = "unknown"

	promBuildInfoGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "kask_build_info",
			Help:        "Build information",
			ConstLabels: map[string]string{"version": version, "git": gitTag, "build_date": buildDate, "build_host": buildHost, "go_version": runtime.Version()},
		})
)

func init() {
	prometheus.MustRegister(promHTTPReqsCounterVec, promDurationHistoVec, promBuildInfoGauge)
	promBuildInfoGauge.Set(1)
}

func main() {
	flag.Parse()

	config, err := ReadConfig(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := NewLogger(os.Stdout, config.ServiceName, config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Initializing Kask %s (Git: %s, Go version: %s, Build host: %s, Timestamp: %s)...", version, gitTag, runtime.Version(), buildHost, buildDate)

	store, err := NewCassandraStore(config)
	if err != nil {
		logger.Fatal("Error connecting to Cassandra: %s", err)
		os.Exit(1)
	}

	// Close the database connection before returning from main()
	defer store.Close()

	// Kask CRUD operations
	handler := &HTTPHandler{store, config, logger}

	// Wrap in middlewares
	dispatcher := ValidatingKeyParserMiddleware(config.BaseURI, handler)
	dispatcher = PrometheusInstrumentationMiddleware(promHTTPReqsCounterVec, promDurationHistoVec, dispatcher)

	http.Handle(config.BaseURI, dispatcher)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/healthz", http.HandlerFunc(Healthz))

	// Serve OpenAPI specification (if so-configured).
	if config.OpenAPISpec != "" {
		http.Handle("/openapi", OpenAPI(config, logger))
	}

	listen := fmt.Sprintf("%s:%d", config.Address, config.Port)

	// TLS configuration
	if config.TLS.CertPath != "" {
		logger.Info("Starting service as https://%s%s", listen, config.BaseURI)
		log.Fatal(http.ListenAndServeTLS(listen, config.TLS.CertPath, config.TLS.KeyPath, nil))
	} else {
		logger.Info("Starting service as http://%s%s", listen, config.BaseURI)
		log.Fatal(http.ListenAndServe(listen, nil))
	}
}
