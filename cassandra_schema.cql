
;; Sample schema

CREATE KEYSPACE kask WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
CREATE TABLE kask.values (key text PRIMARY KEY, value blob);
