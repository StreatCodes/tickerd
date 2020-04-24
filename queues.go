package main

import "github.com/jmoiron/sqlx"

//Queue mapped to DB type
type Queue struct {
	ID    int64
	Name  string
	Email string
}

func createQueue(db *sqlx.DB, name, email string) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO Queue(Name, Email) VALUES (?, ?)`,
		name, email,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
