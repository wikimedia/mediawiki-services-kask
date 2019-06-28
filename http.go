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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gocql/gocql"
)

type contextKey int

const kaskKey contextKey = iota

// Problem corresponds to an HTTP problem (https://tools.ietf.org/html/rfc7807)
type Problem struct {
	Code     int    `json:"-"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

// BadRequest is an HTTP problem (RFC7807) corresponding to a status 400 response.
func BadRequest(instance string) Problem {
	return Problem{
		Code:     400,
		Type:     "https://www.mediawiki.org/wiki/Kask/errors/bad_request",
		Title:    "Bad request",
		Detail:   "The request was incorrect or malformed",
		Instance: instance,
	}
}

// NotAuthorized is an HTTP problem (RFC7807) corresponding to a status 401 response.
func NotAuthorized(instance string) Problem {
	return Problem{
		Code:     401,
		Type:     "https://www.mediawiki.org/wiki/Kask/errors/not_authorized",
		Title:    "Not authorized",
		Detail:   "Unable to authorize request",
		Instance: instance,
	}
}

// NotFound is an HTTP problem (RFC7807) corresponding to a status 404 response.
func NotFound(instance string) Problem {
	return Problem{
		Code:     404,
		Type:     "https://www.mediawiki.org/wiki/Kask/errors/not_found",
		Title:    "Not found",
		Detail:   "The value you requested was not found",
		Instance: instance,
	}
}

// InternalServerError is an HTTP problem (RFC7807) corresponding to a status 500 response.
func InternalServerError(instance string) Problem {
	return Problem{
		Code:     500,
		Type:     "https://www.mediawiki.org/wiki/Kask/errors/server_error",
		Title:    "Internal server error",
		Detail:   "The server encountered an error with your request",
		Instance: instance,
	}
}

// HTTPError applies an HTTP problem to an HTTP response
func HTTPError(w http.ResponseWriter, p Problem) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(p.Code)
	j, err := json.MarshalIndent(p, "", "  ")

	if err != nil {
		fmt.Fprintln(w, "UNABLE TO MARSHALL JSON ERROR RESPONSE; THIS IS A BUG!")
		return
	}
	fmt.Fprintln(w, string(j))
}

// getRequestID returns the value of an `X-Request-ID` header when present, or a default otherwise.
func getRequestID(r *http.Request) string {
	id := r.Header.Get("X-Request-ID")

	// Sets a default for request id when it is not forwarded in the X-Request-ID header
	if id == "" {
		return "00000000-0000-0000-0000-000000000000"
	}

	return id
}

// HTTPHandler encapsulates the Kask request handlers and their dependencies.
type HTTPHandler struct {
	store  Store
	config *Config
	log    *Logger
}

// ServeHTTP accepts requests (of the base URI) for any HTTP method, and dispatches them to the appropriate handler.
func (env *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		env.get(w, r)
	case http.MethodPost:
		env.post(w, r)
	case http.MethodPut:
		env.put(w, r)
	case http.MethodDelete:
		env.delete(w, r)
	default:
		HTTPError(w, BadRequest(r.URL.Path))
		env.log.RequestID(getRequestID(r)).Log(LogError, "Unsupported HTTP method used: (%s)", r.Method)
	}
}

// GET requests
func (env *HTTPHandler) get(w http.ResponseWriter, r *http.Request) {
	key := r.Context().Value(kaskKey).(string)
	value, err := env.store.Get(key)
	if err != nil {
		if err == gocql.ErrNotFound {
			HTTPError(w, NotFound(r.URL.Path))
		} else {
			HTTPError(w, InternalServerError(r.URL.Path))
			env.log.RequestID(getRequestID(r)).Log(LogError, "Error reading from storage (%v)", err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")

	if _, err := w.Write(value.Value); err != nil {
		env.log.RequestID(getRequestID(r)).Log(LogError, "Error writing HTTP response body: (%s)", err)
	}
}

// POST requests
func (env *HTTPHandler) post(w http.ResponseWriter, r *http.Request) {
	key := r.Context().Value(kaskKey).(string)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		HTTPError(w, InternalServerError(r.URL.Path))
		env.log.RequestID(getRequestID(r)).Log(LogDebug, "Error reading body of POST request: (%s)", err)
		return
	}

	if len(body) == 0 {
		HTTPError(w, BadRequest(r.URL.Path))
		env.log.RequestID(getRequestID(r)).Log(LogError, "Request body is empty")
		return
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err := env.store.Set(key, body, env.config.DefaultTTL); err != nil {
		HTTPError(w, InternalServerError(r.URL.Path))
		env.log.RequestID(getRequestID(r)).Log(LogError, "Error writing to storage (%v)", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/octet-stream")
}

// PUT requests
func (env *HTTPHandler) put(w http.ResponseWriter, r *http.Request) {
	HTTPError(w, BadRequest(r.URL.Path))
}

// DELETE requests
func (env *HTTPHandler) delete(w http.ResponseWriter, r *http.Request) {
	key := r.Context().Value(kaskKey).(string)
	if err := env.store.Delete(key); err != nil {
		HTTPError(w, InternalServerError(r.URL.Path))
		env.log.RequestID(getRequestID(r)).Log(LogError, "Error deleting in storage (%v)", err)
	}
	w.WriteHeader(http.StatusNoContent)
}

// ValidatingKeyParserMiddleware returns HTTP middleware that parses a key from the remaining URI, and adds it to
// the request context.
func ValidatingKeyParserMiddleware(baseURI string, next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := strings.Split(r.URL.Path, baseURI)[1:]

		if len(base) == 0 {
			HTTPError(w, NotFound(r.URL.Path))
			return
		}

		// Checks if there are queries in URL
		if len(r.URL.RawQuery) > 0 {
			HTTPError(w, BadRequest(r.URL.Path))
			return
		}

		list := strings.Split(base[0], "/")

		// Checks if there are more than one key passed in in the URL after the baseURI
		if len(list) > 1 {
			HTTPError(w, NotFound(r.URL.Path))
			return
		}

		key := list[0]

		if key == "" {
			HTTPError(w, NotFound(r.URL.Path))
			return
		}

		ctx := context.WithValue(r.Context(), kaskKey, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// statusObserver wraps the existing ResponseWriter in order to track the code for later use in categorizing metrics.
type statusObserver struct {
	http.ResponseWriter
	status int
}

// WriteHeader writes the HTTP response status code to the ResponseWriter and status.
func (r *statusObserver) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// PrometheusInstrumentationMiddleware is a middleware that wraps the provided http.Handler
// to count and observe the request and its duration with the provided CounterVec and HistogramVec.
func PrometheusInstrumentationMiddleware(counter *prometheus.CounterVec, obs *prometheus.HistogramVec, next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		// Sets the response status to the default 200 for calls that do not call WriteHeader
		d := statusObserver{w, 200}

		next.ServeHTTP(&d, r)

		obs.WithLabelValues(strconv.Itoa(d.status), r.Method).Observe(time.Since(now).Seconds())
		counter.WithLabelValues(strconv.Itoa(d.status), r.Method).Inc()
	})
}

// Healthz is an HTTP handler function that simply returns 200; Healthz serves as a
// readiness test for Kubernetes.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// OpenAPI is an HTTP handler function that serves an OpenAPI specfication file.  The file is assumed to
// be in YAML format, and can be templated to include the configured base URI in path statements.
func OpenAPI(config *Config, logger *Logger) http.HandlerFunc {
	tmpl, err := template.New(path.Base(config.OpenAPISpec)).ParseFiles(config.OpenAPISpec)
	if err != nil {
		logger.Error("Unable to parse %s as template: %s", config.OpenAPISpec, err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			HTTPError(w, InternalServerError(r.URL.Path))
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		if exerr := tmpl.Execute(w, config); exerr != nil {
			logger.Error("Unable to template OpenAPI spec: %s", exerr)
		}
	})
}
