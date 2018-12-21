package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

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
