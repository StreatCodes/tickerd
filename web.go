package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var sockets = make(map[int]*websocket.Conn)
var handlers = make(map[string]WSHandler)

//WSHandler receives json as bytes
//Returns a JSON encoded result or throws an error
type WSHandler func(userID int, reqJSON []byte) ([]byte, error)

func echo(userID int, reqJSON []byte) ([]byte, error) {
	return reqJSON, nil
}

func registerHandler(name string, handler WSHandler) {
	handlers[name] = handler
}

func createFileHandler(frontendDir http.Dir) func(w http.ResponseWriter, r *http.Request) {
	fs := http.FileServer(frontendDir)

	return func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean(r.URL.Path)
		f, err := frontendDir.Open(p)

		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Open("./frontend/index.html")
			if err != nil {
				http.Error(w, "Error opening index.html - "+err.Error(), http.StatusInternalServerError)
			}
			_, err = io.Copy(w, f)
			if err != nil {
				http.Error(w, "Error writing index.html to response - "+err.Error(), http.StatusInternalServerError)
			}
			f.Close()
		} else {
			f.Close()
			fs.ServeHTTP(w, r)
		}
	}
}

func initWeb() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/ws", initWebsocket)
	http.HandleFunc("/", createFileHandler(http.Dir("frontend")))

	httpAddr := ":8080"
	fmt.Printf("Starting server at %s\n", httpAddr)
	log.Fatal(http.ListenAndServe(httpAddr, nil))
}

func initWebsocket(w http.ResponseWriter, r *http.Request) {
	//Check user is authenticated
	token := r.URL.Query().Get("token")
	btoken, err := hex.DecodeString(token)
	if err != nil {
		http.Error(w, "Failed to decode token", http.StatusBadRequest)
		return
	}

	user, err := validateToken(btoken)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "User validation failed", http.StatusUnauthorized)
		return
	}

	//Auth passed, upgrade user
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sockets[user.ID] = conn
	go websocketHandler(user.ID, conn)
}

//WSReq a message received from a websocket connection
type WSReq struct {
	ID     uint
	Method string
	Params json.RawMessage
}

//WSResp the response from a WSReq
type WSResp struct {
	ID     uint
	Result json.RawMessage
	Error  *string
}

//This can be multithreaded per message
func websocketHandler(userID int, conn *websocket.Conn) {
	var err error
	responseChan := make(chan WSResp)

	//Send responses back to client
	go func() {
		for response := range responseChan {
			err := conn.WriteJSON(response)
			if err != nil {
				fmt.Printf("Failed to write json response %s\n", err)
				close(responseChan)
			}
		}
	}()

	for {
		//Get next avialable message
		var reader io.Reader
		_, reader, err = conn.NextReader()
		if err != nil {
			err = fmt.Errorf("Error getting next WS reader - %s", err)
			break
		}

		var req WSReq
		dec := json.NewDecoder(reader)
		err = dec.Decode(&req)
		if err != nil {
			err = fmt.Errorf("Error decoding message - %s", err)
			break
		}

		//Handle request
		go handleRequest(userID, req, responseChan)
	}

	if !strings.Contains(err.Error(), "close 1001") {
		fmt.Printf("Websocket error, closing connection: %s\n", err)
		conn.Close()
	}
	close(responseChan)
}

func handleRequest(userID int, req WSReq, responseChan chan WSResp) {
	var err error
	res := WSResp{
		ID: req.ID,
	}

	// method string, params json.RawMessage
	if t, ok := handlers[req.Method]; ok {
		res.Result, err = t(userID, req.Params)
	} else {
		err = fmt.Errorf("Handler %s not found", req.Method)
	}

	if err != nil {
		tmp := err.Error()
		res.Error = &tmp
	}

	responseChan <- res
}
