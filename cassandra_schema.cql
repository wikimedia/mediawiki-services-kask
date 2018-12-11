
;; -*- mode:sql; -*-

DROP KEYSPACE IF EXISTS kask_test_keyspace;
CREATE KEYSPACE kask_test_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
CREATE TABLE kask_test_keyspace.test_table (key text PRIMARY KEY, value blob);
