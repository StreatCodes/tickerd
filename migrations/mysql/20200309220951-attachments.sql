-- +migrate Up
CREATE TABLE Attachment (
	ReplyID INT NOT NULL,
	HASH BINARY(32),
	OriginalName TEXT NOT NULL,

	FOREIGN KEY (ReplyID) REFERENCES Reply(ID)
);

-- +migrate Down
DROP TABLE Attachment;