-- +migrate Up
CREATE TABLE test_hello(id int);

-- +migrate Down
DROP TABLE test_hello;
