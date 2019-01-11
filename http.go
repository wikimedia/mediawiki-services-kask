package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gocql/gocql"
)

type contextKey int

const kaskKey contextKey = iota

type Problem struct {
	Code     int    `json:"-"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

func BadRequest(instance string) Problem {
	return Problem{
		Code:     400,
		Type:     "https://www.mediawiki.org/wiki/probs/bad-request",
		Title:    "Bad request",
		Detail:   "The request was incorrect or malformed",
		Instance: instance,
	}
}

func NotAuthorized(instance string) Problem {
	return Problem{
		Code:     401,
		Type:     "https://www.mediawiki.org/wiki/probs/not-authorized",
		Title:    "Not authorized",
		Detail:   "Unable to authorize request",
		Instance: instance,
	}
}

func NotFound(instance string) Problem {
	return Problem{
		Code:     404,
		Type:     "https://www.mediawiki.org/wiki/probs/not-found",
		Title:    "Not found",
		Detail:   "The value you requested was not found",
		Instance: instance,
	}
}

func InternelServerError(instance string) Problem {
	return Problem{
		Code:     500,
		Type:     "https://www.mediawiki.org/wiki/probs/internal-server-error",
		Title:    "Internal server error",
		Detail:   "The server encountered an error with your request",
		Instance: instance,
	}
}

func HttpError(w http.ResponseWriter, p Problem) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(p.Code)
	j, err := json.MarshalIndent(p, "", "  ")
	// FIXME: Think about proper errorhandling
	if err != nil {
		log.Printf("Oh noes; Failed to encode problem as JSON: %s", err)
	}
	fmt.Fprintln(w, string(j))
}

type HttpHandler struct {
	store  Store
	config *Config
	log    *Logger
}

func (env *HttpHandler) Dispatch(w http.ResponseWriter, r *http.Request) {
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
		HttpError(w, BadRequest(r.URL.Path))
	}
}

func (env *HttpHandler) get(w http.ResponseWriter, r *http.Request) {
	key := r.Context().Value(kaskKey).(string)
	value, err := env.store.Get(key)
	if err != nil {
		if err == gocql.ErrNotFound {
			HttpError(w, NotFound(r.URL.Path))
		} else {
			HttpError(w, InternelServerError(r.URL.Path))
		}
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value.Value)
}

func (env *HttpHandler) post(w http.ResponseWriter, r *http.Request) {
	key := r.Context().Value(kaskKey).(string)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		HttpError(w, InternelServerError(r.URL.Path))
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err := env.store.Set(key, body, env.config.DefaultTTL); err != nil {
		env.log.Error("Unable to persist value (%s)", err)
		HttpError(w, InternelServerError(r.URL.Path))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/octet-stream")
}

func (env *HttpHandler) put(w http.ResponseWriter, r *http.Request) {
	HttpError(w, BadRequest(r.URL.Path))
}

func (env *HttpHandler) delete(w http.ResponseWriter, r *http.Request) {
	key := r.Context().Value(kaskKey).(string)
	if err := env.store.Delete(key); err != nil {
		HttpError(w, InternelServerError(r.URL.Path))
	}
	w.WriteHeader(http.StatusNoContent)
}

// NewParseKeyMiddleware is a function that accepts a prefix, and returns HTTP middleware that
// parses a key from the remaining URI.
func NewParseKeyMiddleware(baseURI string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			base := strings.Split(r.URL.Path, baseURI)[1:]

			if len(base) == 0 {
				HttpError(w, NotFound(r.URL.Path))
				return
			}

			// Checks if there are queries in URL
			if len(r.URL.RawQuery) > 0 {
				HttpError(w, BadRequest(r.URL.Path))
				return
			}

			list := strings.Split(base[0], "/")

			// Checks if there are more than one key passed in in the URL after the baseURI
			if len(list) > 1 {
				HttpError(w, NotFound(r.URL.Path))
				return
			}

			key := list[0]
			ctx := context.WithValue(r.Context(), kaskKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
