-- +migrate Up
CREATE TABLE User(
	ID INT NOT NULL AUTO_INCREMENT,
	Name  TEXT NOT NULL,
	Email TEXT NOT NULL,
	Admin BOOLEAN,
	
	PRIMARY KEY (ID)
);

-- +migrate Down
DROP TABLE User;
