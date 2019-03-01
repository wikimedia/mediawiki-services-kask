
# Copyright 2019 Clara Andrew-Wani <candrew@wikimedia.org>, Eric Evans <eevans@wikimedia.org>,
# and Wikimedia Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


GOPATH ?= /usr/share/gocode
GOTEST_ARGS ?= 
CONFIG ?= config.yaml.test
GO_PACKAGES := ./...


build:
	GOPATH=$(GOPATH) go build kask.go config.go http.go logging.go storage.go

functional-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=functional -config $(CONFIG)

unit-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=unit

integration-test: build
	GOPATH=$(GOPATH) go test $(GOTEST_ARGS) -tags=integration -config $(CONFIG)

check:
	@if [ -n "`goimports -l *.go`" ]; then \
	    echo "goimports: format errors detected" >&2; \
	    false; \
	fi
	@if [ -n "`gofmt -l *.go`" ]; then \
	    echo "gofmt: format errors detected" >&2; \
	    false; \
	fi
	golint -set_exit_status $(GO_PACKAGES)
	go vet $(GO_PACKAGES)

test: unit-test check

clean:
	rm -f kask

.PHONY: build functional-test unit-test integration-test check test clean
