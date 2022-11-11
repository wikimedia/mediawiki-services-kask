//go:build integration
// +build integration

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
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestKask(t *testing.T) {
	config, err := ReadConfig(*confFile)

	if err != nil {
		t.Error(err)
		return
	}

	key := RandString(8)
	value := RandString(32)
	url := fmt.Sprintf("http://%s:%d%s%s", config.Address, config.Port, config.BaseURI, key)

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
}
