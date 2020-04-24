package main

import (
	"crypto/rand"
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

//SocketID used to keep track of websockets
type SocketID [16]byte

var sockets = make(map[SocketID]*websocket.Conn)
var handlers = make(map[string]WSHandler)

//WSHandler receives json as bytes
//Returns a JSON encoded result or throws an error
type WSHandler func(reqJSON []byte) ([]byte, error)

func registerHandler(name string, handler WSHandler) {
	handlers[name] = handler
}

//Serves index.html if the path isn't found
func createFileHandler(frontendPath string) func(w http.ResponseWriter, r *http.Request) {
	frontendDir := http.Dir(frontendPath)
	fs := http.FileServer(frontendDir)

	return func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean(r.URL.Path)
		f, err := frontendDir.Open(p)

		if errors.Is(err, os.ErrNotExist) {
			f, err := os.Open(path.Join(frontendPath, "index.html"))
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

//This seems ridiculous
func generateSocketID() SocketID {
	token := SocketID{}
	tempToken := make([]byte, len(SocketID{}))
	rand.Read(tempToken)

	for i := range tempToken {
		token[i] = tempToken[i]
	}

	return token
}

func initWeb() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/ws", initWebsocket)
	http.HandleFunc("/", createFileHandler("./frontend"))

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

	socketID := generateSocketID()
	sockets[socketID] = conn
	go websocketHandler(socketID, user.ID, conn)
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
func websocketHandler(socketID SocketID, userID int, conn *websocket.Conn) {
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
	delete(sockets, socketID)
}

func handleRequest(userID int, req WSReq, responseChan chan WSResp) {
	var err error
	res := WSResp{
		ID: req.ID,
	}

	// method string, params json.RawMessage
	if t, ok := handlers[req.Method]; ok {
		res.Result, err = t(req.Params)
	} else {
		err = fmt.Errorf("Handler %s not found", req.Method)
	}

	if err != nil {
		tmp := err.Error()
		res.Error = &tmp
	}

	responseChan <- res
}
