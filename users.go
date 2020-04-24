package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

//User the data describing a user
type User struct {
	ID    int64  `db:"ID"`
	Name  string `db:"Name"`
	Email string `db:"Email"`
	Admin bool   `db:"Admin"`
}

func createUser(name, email string, admin bool) (User, error) {
	res := DB.MustExec(
		`INSERT INTO User (Name, Email, Admin) VALUES (?, ?, ?)`,
		name, email, admin,
	)

	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}

	return User{
		ID:    id,
		Name:  name,
		Email: email,
		Admin: admin,
	}, nil
}

//refine this
func tryLogin(email, password string) (int64, bool, error) {
	type userInfo struct {
		ID       int64  `db:"ID"`
		Password []byte `db:"Password"`
	}
	user := userInfo{}

	err := DB.Get(&user, `SELECT ID, Password FROM User WHERE Email=?`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, fmt.Errorf("No user with that email")
	} else if err != nil {
		return 0, false, err
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return 0, false, err
	}

	return user.ID, true, nil
}

func createSession(userID int64) ([]byte, error) {
	tokenLength := 128
	token := make([]byte, tokenLength)
	c, err := rand.Read(token)
	if err != nil {
		return nil, err
	}

	if c < 128 {
		return nil, fmt.Errorf("Error couldn't read the full %d bytes of random data", tokenLength)
	}

	res := DB.MustExec(`INSERT INTO Session (UserID, Token) VALUES (?, ?)`, userID, token)
	_, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return token, nil
}

func validateToken(token []byte) (User, error) {
	var user User

	err := DB.Get(&user, `SELECT User.ID, User.Name, User.Email, User.Admin
		FROM Session
		JOIN User ON User.ID = Session.UserID
		WHERE Session.Token=?`, token)

	return user, err
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	type loginRequest struct {
		Email    string
		Password string
	}

	var loginReq loginRequest
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := dec.Decode(&loginReq)
	if err != nil {
		http.Error(w, "Error decoding request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID, ok, err := tryLogin(loginReq.Email, loginReq.Password)
	if err != nil {
		http.Error(w, "An error occured while attempting to log you in - "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Login request failed. Email or Password were incorrect.", http.StatusUnauthorized)
		return
	}

	btoken, err := createSession(userID)
	token := hex.EncodeToString(btoken)

	enc := json.NewEncoder(w)
	err = enc.Encode(token)
	if err != nil {
		fmt.Println("Error encoding response - " + err.Error())
	}
}

func wshMe(userID int64, body []byte) ([]byte, error) {
	var user User
	err := DB.Get(&user, `SELECT ID, Name, Email, Admin FROM User where ID=?`, userID)
	if err != nil {
		return nil, err
	}

	resp, err := json.Marshal(user)
	return resp, err
}
