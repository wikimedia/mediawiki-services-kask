kask
====

[Kask][wiki page] is a (multi-master) replicated key-value storage service.

## Building

Kask's dependencies target packages shipped in Debian Stretch; To build and
execute on a Debian machine:

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

#### Executing Tests and Benchmarks

    $ make GOTEST_ARGS="-bench=. -benchmem" test

#### Experimental integration tests

This branch includes experimental support for Javascript integration
tests using a [nodejs module](https://github.com/eevans/api-testing).

    $ TEST_URL=http://localhost:8080/v1 node_modules/api-testing/bin/runner .api-testing.yaml quick
    All good (4/4 tests passed)


*NOTE: `config.yaml.test` is excluded from version control and is recommended for local configuration.*

## Running

Create the Cassandra schema

    $ cqlsh -f cassandra_schema.cql

Startup

    $ ./kask --config <config file>

## Using

    $ curl -X POST -H 'Content-Type: application/octet-stream' \
           -d 'sample value' http://api.example.org/sessions/v1/test_key
    HTTP/1.1 201 CREATED
    Content-Type: application/octet-stream
    Date: Tue, 11 Dec 2018 22:50:46 GMT
    Content-Length: 0

    $ curl http://api.example.org/sessions/v1/test_key
    HTTP/1.1 200 OK
    Content-Type: application/octet-stream
    Date: Tue, 11 Dec 2018 22:51:10 GMT
    Content-Length: 3

    sample value

## See also

For more information about Kask, see the [wiki page].

[wiki page]: https://www.mediawiki.org/wiki/Kask
