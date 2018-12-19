package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

const root string = "/sessions/v1/"

func Getenv(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	} else {
		return value
	}
}

func getHandler(w http.ResponseWriter, r *http.Request, store *Store, key string) {
	value, err := store.Get(key)
	if err != nil {
		// FIXME: This needs to differentiate between a failure to execute the SELECT, and
		// record that is not found (and should ultimately 404).
		log.Printf("Error reading key (%s)", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value)
}

func postHandler(w http.ResponseWriter, r *http.Request, store *Store, key string) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err := store.Set(key, body); err != nil {
		log.Printf("Error setting value (%s)", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
}

func putHandler(w http.ResponseWriter, r *http.Request, store *Store, key string) {
	log.Println(key)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, store *Store, key string) {
	if err := store.Delete(key); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// FIXME: Not good enough; Will violate Element of Least Surprise if
// base contains more than one path element (e.g. 'a/b/c/...').
func parseKey(url string) (string, error) {
	base := strings.Replace(url, root, "", 1)
	if base == "" {
		return base, errors.New("No key found")
	}
	return path.Base(url), nil
}

func dispatch(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key, err := parseKey(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			getHandler(w, r, s, key)
		case http.MethodPost:
			postHandler(w, r, s, key)
		case http.MethodPut:
			putHandler(w, r, s, key)
		case http.MethodDelete:
			deleteHandler(w, r, s, key)
		default:
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	}
}

func main() {
	hostname := Getenv("CASSANDRA_HOST", "localhost")
	port := Getenv("CASSANDRA_PORT", "9042")
	keyspace := Getenv("CASSANDRA_KEYSPACE", "kask_test_keyspace")
	table := Getenv("CASSANDRA_TABLE", "test_table")

	portNum, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("%s is not a valid TCP port number!", port)
	}

	// TODO: Handle errors...
	store, _ := NewStore(hostname, portNum, keyspace, table)

	http.HandleFunc(root, dispatch(store))
	log.Fatal(http.ListenAndServe(":8080", nil))

	defer store.Close()
}
