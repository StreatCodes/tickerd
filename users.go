package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/asdine/storm/v3"
	"golang.org/x/crypto/bcrypt"
)

//User the data describing a user
type User struct {
	ID       int64 `storm:"increment"`
	Name     string
	Email    string `storm:"unique"`
	Password []byte `json:"-"`
	Admin    bool
}

//Session websocket info
type Session struct {
	Token  []byte `storm:"id"`
	UserID int64
}

func createUser(name, email string, password []byte, admin bool) (User, error) {
	var user User
	encryptedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return user, err
	}

	user.Name = name
	user.Email = email
	user.Admin = admin
	user.Password = encryptedPassword

	err = tickerDB.Save(&user)
	if err != nil {
		return user, err
	}

	return user, nil
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

	session := Session{
		Token:  token,
		UserID: userID,
	}
	err = tickerDB.Save(&session)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func validateToken(token []byte) (User, error) {
	var session Session
	var user User

	err := tickerDB.Get("Session", token, &session)
	if err != nil {
		return user, err
	}

	err = tickerDB.Get("User", session.UserID, &user)
	if err != nil {
		return user, err
	}

	return user, nil
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

	var user User
	err = tickerDB.One("Email", loginReq.Email, &user)
	if err == storm.ErrNotFound {
		http.Error(w, "Login request failed. Email or Password were incorrect.", http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Println("Error looking up user: " + err.Error())
		http.Error(w, "An error occured while attempting to log you in", http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(loginReq.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		http.Error(w, "Login request failed. Email or Password were incorrect.", http.StatusUnauthorized)
		return
	} else if err != nil {
		fmt.Println("Error occurred when comparing password hash: " + err.Error())
		http.Error(w, "An error occured while attempting to log you in", http.StatusInternalServerError)
		return
	}

	btoken, err := createSession(user.ID)
	token := hex.EncodeToString(btoken)

	enc := json.NewEncoder(w)
	err = enc.Encode(token)
	if err != nil {
		fmt.Println("Error encoding response - " + err.Error())
	}
}
