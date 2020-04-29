package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
)

//Queue mapped to DB type
type Queue struct {
	ID    int64  `storm:"increment"`
	Name  string `storm:"unique"`
	Email string `storm:"unique"`
}

//Validate queue fields
func (q *Queue) validate() error {
	if q.Name == "" {
		return errors.New("Name can't be blank")
	}

	_, err := mail.ParseAddress(q.Email)
	if err != nil {
		return fmt.Errorf("Invalid Email address: %s", err)
	}

	return nil
}

//WSCreateQueue websocket handler for the creation of queues
func WSCreateQueue(userID int64, reqJSON []byte) ([]byte, error) {
	var queue Queue
	err := json.Unmarshal(reqJSON, &queue)
	if err != nil {
		return nil, err
	}

	err = queue.validate()
	if err != nil {
		return nil, fmt.Errorf("Error validating new queue: %s", err)
	}

	err = tickerDB.Save(&queue)
	if err != nil {
		return nil, fmt.Errorf("Error saving queue: %s", err)
	}

	res, err := json.Marshal(queue)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//WSUpdateQueue websocket handler for the updating of queues
func WSUpdateQueue(userID int64, reqJSON []byte) ([]byte, error) {
	var queue Queue
	err := json.Unmarshal(reqJSON, &queue)
	if err != nil {
		return nil, err
	}

	exists, err := tickerDB.KeyExists("Queue", queue.ID)
	if err != nil {
		return nil, fmt.Errorf("Error looking up queue: %s", err)
	}

	if !exists {
		return nil, fmt.Errorf("Queue does not exist")
	}

	err = queue.validate()
	if err != nil {
		return nil, fmt.Errorf("Error validating updated queue: %s", err)
	}

	err = tickerDB.Save(&queue)
	if err != nil {
		return nil, fmt.Errorf("Error saving queue: %s", err)
	}

	res, err := json.Marshal(queue)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//WSListQueues websocket handler for listing queues
func WSListQueues(userID int64, reqJSON []byte) ([]byte, error) {
	var queues []Queue
	err := tickerDB.All(&queues)
	if err != nil {
		return nil, err
	}

	res, err := json.Marshal(queues)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//WSDeleteQueue websocket handler for the deletion of queues
func WSDeleteQueue(userID int64, reqJSON []byte) ([]byte, error) {
	var queueID int64
	err := json.Unmarshal(reqJSON, &queueID)
	if err != nil {
		return nil, err
	}

	exists, err := tickerDB.KeyExists("Queue", queueID)
	if err != nil {
		return nil, fmt.Errorf("Error looking up queue: %s", err)
	}

	if !exists {
		return nil, fmt.Errorf("Queue does not exist")
	}

	err = tickerDB.DeleteStruct(&Queue{ID: queueID})
	if err != nil {
		return nil, fmt.Errorf("Error deleting queue: %s", err)
	}

	return nil, nil
}
