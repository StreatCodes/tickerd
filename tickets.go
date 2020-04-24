package main

import (
	"encoding/json"
	"time"
)

const (
	//NEW ticket
	NEW = iota
	//OPEN ticket
	OPEN = iota
	//CLOSED ticket
	CLOSED = iota
)

//Ticket go representation of DB type
type Ticket struct {
	ID        int64
	QueueID   int64
	Subject   string
	Status    int64
	Priority  int64
	CreatedAt time.Time
}

const (
	//REPLY correspondence on ticket
	REPLY = iota
	//COMMENT correspondence on ticket
	COMMENT = iota
)

type Reply struct {
	Body       string
	Type       string
	RenderType string
}

//WSCreateTicket websocket handler for the creation of tickets,
//It also writes the first reply (the body of the original ticket)
func WSCreateTicket(userID int64, reqJSON []byte) ([]byte, error) {
	type ticketReq struct {
		Ticket Ticket
		Reply  Reply
	}

	type ticketRes struct {
		TicketID int64
		ReplyID  int64
	}

	var ticket ticketReq

	err := json.Unmarshal(reqJSON, &ticket)
	if err != nil {
		return nil, err
	}

	//Create ticket in DB
	tx, err := DB.Begin()
	row, err := tx.Exec(
		`INSERT INTO 
		Ticket(QueueID, Subject, Status, Priority, CreatedAt)
		VALUES (?, ?, ?, 0, NOW())`,
		ticket.Ticket.QueueID, ticket.Ticket.Subject, NEW,
	)
	if err != nil {
		return nil, err
	}

	ticketID, err := row.LastInsertId()
	if err != nil {
		return nil, err
	}

	//Create reply in DB
	row, err = tx.Exec(
		`INSERT INTO 
		Ticket(TicketID, Body, Type, RenderType, CreatedAt)
		VALUES (?, ?, ?, ?, NOW())`,
		ticketID, ticket.Reply.Body, ticket.Reply.Type, ticket.Reply.RenderType,
	)
	if err != nil {
		return nil, err
	}

	replyID, err := row.LastInsertId()
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//Response json
	res, err := json.Marshal(ticketRes{TicketID: ticketID, ReplyID: replyID})
	if err != nil {
		return nil, err
	}

	return res, nil
}
