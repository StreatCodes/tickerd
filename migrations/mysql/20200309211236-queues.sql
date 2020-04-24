-- +migrate Up
CREATE TABLE Queue(
	ID INT NOT NULL AUTO_INCREMENT,
	Name TEXT NOT NULL,
	Email TEXT,

	PRIMARY KEY (ID)
);

-- +migrate Down
DROP TABLE Queue;
