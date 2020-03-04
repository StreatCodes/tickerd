-- +migrate Up
ALTER TABLE User
ADD Password BINARY(64);

-- +migrate Down
ALTER TABLE User
DROP COLUMN Password;