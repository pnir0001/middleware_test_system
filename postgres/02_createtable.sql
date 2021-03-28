\c test_postgres_db;

CREATE TABLE ids
(
    id  CHAR(36) NOT NULL,
    timestamp   bigint,
PRIMARY KEY (id)
);