package main

import (
	"encoding/json"
)

//Queue mapped to DB type
type Queue struct {
	ID    int64 `storm:"increment"`
	Name  string
	Email string
}

//WSCreateQueue websocket handler for the creation of queues
func WSCreateQueue(userID int64, reqJSON []byte) ([]byte, error) {
	var queue Queue
	err := json.Unmarshal(reqJSON, &queue)
	if err != nil {
		return nil, err
	}

	err = tickerDB.Save(&queue)
	if err != nil {
		return nil, err
	}

	res, err := json.Marshal(queue)
	if err != nil {
		return nil, err
	}

	return res, nil
}
