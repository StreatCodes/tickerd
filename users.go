package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

//User the data describing a user
type User struct {
	ID    int
	Name  string
	Email string
	Admin bool
}

//refine this
func tryLogin(email, password string) (int, bool, error) {
	type userInfo struct {
		ID       int
		Password []byte
	}
	var users []userInfo

	err := DB.Select(&users, `SELECT ID, Password from User WHERE Email=$1`, email)
	if err != nil {
		return 0, false, err
	}

	if len(users) < 1 {
		return 0, false, fmt.Errorf("No user with that email")
	}

	user := users[0]
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return 0, false, err
	}

	return user.ID, true, nil
}

func createSession(userID int) ([]byte, error) {
	tokenLength := 128
	token := make([]byte, tokenLength)
	c, err := rand.Read(token)
	if err != nil {
		return nil, err
	}

	if c < 128 {
		return nil, fmt.Errorf("Error couldn't read the full %d bytes of random data", tokenLength)
	}

	res := DB.MustExec(`INSERT INTO Session (UserID, Token) VALUES ($1, $2)`, userID, token)
	_, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return token, nil
}

func validateToken(token []byte) (User, error) {
	var user User

	err := DB.Select(&user, `SELECT User.* FROM Session
		WHERE Session.Token=$1
		JOIN User.ID ON Session.UserID`, token)

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
		http.Error(w, "An error occured while attempting to log you in", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Login request failed. Email or Password were incorrect.", http.StatusUnauthorized)
		return
	}

	token, err := createSession(userID)

	w.Write(token)
}
