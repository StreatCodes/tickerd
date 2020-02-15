package main

import (
	"encoding/json"
	"net/http"
)

//User the data describing a user
type User struct {
	ID    int
	Name  string
	Email string
	Admin bool
}

func tryLogin(email, password string) (bool, error) {

	return true, nil
}

func createSession() ([]byte, error) {
	//TODO
	var token []byte
	return token, nil
}

func validateToken(token []byte) (User, error) {
	//TODO
	return User{ID: 999, Name: "TODO", Email: "TODO", Admin: true}, nil
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

	ok, err := tryLogin(loginReq.Email, loginReq.Password)
	if err != nil {
		http.Error(w, "An error occured while attempting to log you in", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Login request failed. Email or Password were incorrect.", http.StatusUnauthorized)
		return
	}

	token, err := createSession()

	w.Write(token)
}
