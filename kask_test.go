// +build functional

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"testing"
)

func TestKask(t *testing.T) {
	cmd := exec.Command("./kask", "--config", "config.yaml.test")
	config, err := ReadConfig(*confFile)

	if err != nil {
		t.Error(err)
		return
	}

	key := RandString(8)
	value := RandString(32)
	url := fmt.Sprintf("http://%s:%d%s%s", config.Address, config.Port, config.BaseURI, key)

	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	t.Run("404 GET", func(t *testing.T) {
		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			t.Error(err)
			return
		}

		resp, er := http.DefaultClient.Do(req)

		if er != nil {
			t.Error(err)
			return
		}

		AssertEquals(t, http.StatusNotFound, resp.StatusCode, "Incorrect status code returned")
	})

	t.Run("201 POST", func(t *testing.T) {
		req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(value)))

		if err != nil {
			t.Error(err)
			return
		}

		resp, er := http.DefaultClient.Do(req)

		if er != nil {
			t.Error(er)
			return
		}

		AssertEquals(t, http.StatusCreated, resp.StatusCode, "Incorrect status code returned")
	})

	t.Run("200 GET", func(t *testing.T) {
		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			t.Error(err)
			return
		}

		resp, er := http.DefaultClient.Do(req)

		if er != nil {
			t.Error(err)
			return
		}

		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		AssertEquals(t, http.StatusOK, resp.StatusCode, "Incorrect status code returned")

		AssertEquals(t, value, string(body), "Incorrect value returned")
	})

	t.Run("204 DELETE", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", url, nil)

		if err != nil {
			t.Error(err)
			return
		}

		resp, er := http.DefaultClient.Do(req)

		if er != nil {
			t.Error(err)
			return
		}

		AssertEquals(t, http.StatusNoContent, resp.StatusCode, "Incorrect status code returned")
	})

	if err := cmd.Process.Kill(); err != nil {
		log.Fatal("failed to kill process: ", err)
	}
}
