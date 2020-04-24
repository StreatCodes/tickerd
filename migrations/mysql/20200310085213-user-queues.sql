-- +migrate Up
CREATE TABLE UserQueue (
	UserID INT NOT NULL,
	QueueID INT NOT NULL,

	FOREIGN KEY (UserID) REFERENCES User(ID),
	FOREIGN KEY (QueueID) REFERENCES Queue(ID)
);

-- +migrate Down
DROP TABLE UserQueue;