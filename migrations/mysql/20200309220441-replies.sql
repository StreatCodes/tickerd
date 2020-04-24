-- +migrate Up
CREATE TABLE Reply (
	ID INT NOT NULL AUTO_INCREMENT,
	TicketID INT NOT NULL,
	Body INT NOT NULL,
	Type INT NOT NULL, -- Reply / Comment
	RenderType INT NOT NULL, -- HTML / Plain Text
	CreatedAt TIMESTAMP NOT NULL,

	PRIMARY KEY (ID),
	FOREIGN KEY (TicketID) REFERENCES Ticket(ID)
);

-- +migrate Down
DROP TABLE Reply;
