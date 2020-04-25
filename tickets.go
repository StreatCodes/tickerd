package main

import (
	"encoding/json"
	"time"
)

//Ticket go representation of DB type
type Ticket struct {
	ID        int64 `storm:"increment"`
	QueueID   int64 `storm:"index"`
	Subject   string
	Requestor string
	Status    string `storm:"index"`
	Priority  uint8
	CreatedAt time.Time
	Replies   []Reply
	Comments  []Comment
}

//Reply to a Ticket, either from an external email address
//or repliedfrom the web interface
type Reply struct {
	Body        string
	RenderType  string
	Attachments []Attachment
	CreatedAt   time.Time
}

//Comment an internal note on a ticket (does not get sent to the customer)
type Comment struct {
	Body        string
	Attachments []Attachment
	CreatedAt   time.Time
	EditedAt    time.Time
}

//WSCreateTicket websocket handler for the creation of tickets
func WSCreateTicket(userID int64, reqJSON []byte) ([]byte, error) {
	var ticket Ticket
	err := json.Unmarshal(reqJSON, &ticket)
	if err != nil {
		return nil, err
	}

	err = tickerDB.Save(&ticket)
	if err != nil {
		return nil, err
	}

	res, err := json.Marshal(ticket)
	if err != nil {
		return nil, err
	}

	return res, nil
}
