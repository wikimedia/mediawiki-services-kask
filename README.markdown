go-kask
=======

Prototype of Kask in Golang.

First
-----
Create the Cassandra schema

    $ cqlsh -f cassandra_schema.cql

Dependencies used target what has been shipped in Debian Stretch; To build and
execute on a Debian:

    $ apt install golang-github-gocql-gocql-dev
    $ GOPATH=/usr/share/gocode make

If necessary, you can pass environment variables for any of `CASSANDRA_HOST`,
`CASSANDRA_PORT`, `CASSANDRA_KEYSPACE`, or `CASSANDRA_TABLE`.

Then
----

    $ curl -D - -X POST http://localhost:8080/sessions/v1/foo -d 'bar'
    HTTP/1.1 200 OK
    Content-Type: application/octet-stream
    Date: Tue, 11 Dec 2018 22:50:46 GMT
    Content-Length: 0
    
    $ curl -D - -X GET  http://localhost:8080/sessions/v1/foo; echo
    HTTP/1.1 200 OK
    Content-Type: application/octet-stream
    Date: Tue, 11 Dec 2018 22:51:10 GMT
    Content-Length: 3
    
    bar
    $
