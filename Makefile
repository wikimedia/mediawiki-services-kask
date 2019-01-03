
GOPATH ?= /usr/share/gocode

build:
	GOPATH=$(GOPATH) go build kask.go http.go logging.go storage.go

functional-test: build
	GOPATH=$(GOPATH) go test -tags=functional

unit-test: build
	GOPATH=$(GOPATH) go test

test: unit-test functional-test

clean:
	rm -f kask

.PHONY: build functional-test unit-test test clean
