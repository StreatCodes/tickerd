package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/emersion/go-smtp"
)

func initMail() {
	be := &Backend{}

	s := smtp.NewServer(be)

	s.Addr = ":25"
	s.Domain = "streats.dev"
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	log.Println("Starting server at", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// The Backend implements SMTP server methods.
type Backend struct{}

// Login handles a login command with username and password.
func (bkd *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	if username != "streats" || password != "ASDoasodiandS*1dS" {
		return nil, errors.New("Invalid username or password")
	}
	return &MailSession{}, nil
}

// AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (bkd *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return &MailSession{}, nil
}

// A MailSession is returned after successful login.
type MailSession struct{}

func (s *MailSession) Mail(from string, opts smtp.MailOptions) error {
	log.Println("Mail from:", from)
	return nil
}

func (s *MailSession) Rcpt(to string) error {
	log.Println("Rcpt to:", to)
	return nil
}

func (s *MailSession) Data(r io.Reader) error {
	if b, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		log.Println("Data:", string(b))
	}
	return nil
}

func (s *MailSession) Reset() {}

func (s *MailSession) Logout() error {
	return nil
}
