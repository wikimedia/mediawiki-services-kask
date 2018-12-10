package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
)

const root string = "/sessions/v1/"

func getHandler(w http.ResponseWriter, r *http.Request, key string) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write([]byte(fmt.Sprintf("Hello %s!", key)))
}

func postHandler(w http.ResponseWriter, r *http.Request, key string) {
	log.Println(key)
}

func putHandler(w http.ResponseWriter, r *http.Request, key string) {
	log.Println(key)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, key string) {
	log.Println(key)
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

func dispatch(w http.ResponseWriter, r *http.Request) {
	key, err := parseKey(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getHandler(w, r, key)
	case http.MethodPost:
		postHandler(w, r, key)
	case http.MethodPut:
		putHandler(w, r, key)
	case http.MethodDelete:
		deleteHandler(w, r, key)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func main() {
	http.HandleFunc(root, dispatch)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
