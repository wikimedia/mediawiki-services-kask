kask
====

Kask is a (multi-master) replicated key-value storage service.

## Building

Dependencies used target what has been shipped in Debian Stretch; To build and
execute on a Debian:

    $ apt install \
          golang-github-gocql-gocql-dev \
          golang-gopkg-yaml.v2-dev \
          golang-github-prometheus-client-golang-dev \
          golang-golang-x-tools \
          golint \
          git
    $ GOPATH=/usr/share/gocode make

### Executing Tests

    $ make unit-test
    $ CONFIG=config.yaml.test make functional-test

*NOTE: `config.yaml.test` is excluded from version control and is recommended for local configuration.*

## Running

Create the Cassandra schema

    $ cqlsh -f cassandra_schema.cql

Startup

    $ ./kask --config <config file>

## Using

    $ curl -D - -X POST http://localhost:8080/v1/foo -d 'bar'
    HTTP/1.1 201 CREATED
    Content-Type: application/octet-stream
    Date: Tue, 11 Dec 2018 22:50:46 GMT
    Content-Length: 0
    
    $ curl -D - -X GET  http://localhost:8080/v1/foo; echo
    HTTP/1.1 200 OK
    Content-Type: application/octet-stream
    Date: Tue, 11 Dec 2018 22:51:10 GMT
    Content-Length: 3
    
    bar
    $
