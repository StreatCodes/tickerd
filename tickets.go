package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
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

//Validate ticket fields
func (t *Ticket) validate() error {
	exists, err := tickerDB.KeyExists("Queue", t.QueueID)
	if err != nil {
		return fmt.Errorf("Error looking up QueueID: %s", err)
	}

	if !exists {
		return errors.New("Could not find Queue with given QueueID")
	}

	if t.Subject == "" {
		return errors.New("Subject can't be blank")
	}

	_, err = mail.ParseAddress(t.Requestor)
	if err != nil {
		return fmt.Errorf("Invalid Requestor address: %s", err)
	}

	if t.Status == "" {
		t.Status = "new"
	}

	if t.Status != "new" && t.Status != "open" && t.Status != "closed" {
		return errors.New("Status must equal new | open | closed")
	}

	if t.CreatedAt == (time.Time{}) {
		t.CreatedAt = time.Now()
	}

	for i := range t.Replies {
		err = t.Replies[i].validate()
		if err != nil {
			return fmt.Errorf("Reply with index %d is not valid: %s", i, err)
		}
	}

	for i := range t.Comments {
		err = t.Comments[i].validate()
		if err != nil {
			return fmt.Errorf("Comment with index %d is not valid: %s", i, err)
		}
	}

	return nil
}

//Reply to a Ticket, either from an external email address
//or repliedfrom the web interface
type Reply struct {
	Body             string
	RenderType       string
	AttachmentHashes []AttachmentHash
	CreatedAt        time.Time
}

func (r *Reply) validate() error {
	if r.RenderType == "" {
		r.RenderType = "plaintext"
	}

	if r.RenderType != "plaintext" && r.RenderType != "html" {
		return errors.New("RenderType must equal plaintext | html")
	}

	for _, attachment := range r.AttachmentHashes {
		exists, err := tickerDB.KeyExists("Attachment", attachment)
		if err != nil {
			return fmt.Errorf("Error looking up attachment: %s", err)
		}

		if !exists {
			return errors.New("Attachment hash does not exist, upload an attachments first")
		}
	}

	if r.CreatedAt == (time.Time{}) {
		r.CreatedAt = time.Now()
	}

	return nil
}

//Comment an internal note on a ticket (does not get sent to the customer)
type Comment struct {
	Body             string
	AttachmentHashes []AttachmentHash
	CreatedAt        time.Time
	EditedAt         *time.Time
}

func (c *Comment) validate() error {
	for _, attachment := range c.AttachmentHashes {
		exists, err := tickerDB.KeyExists("Attachment", attachment)
		if err != nil {
			return fmt.Errorf("Error looking up attachment: %s", err)
		}

		if !exists {
			return errors.New("Attachment hash does not exist, upload an attachments first")
		}
	}

	if c.CreatedAt == (time.Time{}) {
		c.CreatedAt = time.Now()
	}

	return nil
}

//WSCreateTicket websocket handler for the creation of tickets
func WSCreateTicket(userID int64, reqJSON []byte) ([]byte, error) {
	var ticket Ticket
	err := json.Unmarshal(reqJSON, &ticket)
	if err != nil {
		return nil, err
	}

	//Validate ticket, replies and comments
	err = ticket.validate()
	if err != nil {
		return nil, fmt.Errorf("Error validating new ticket: %s", err)
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
