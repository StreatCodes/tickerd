package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var sockets = make(map[int]*websocket.Conn)
var handlers = make(map[string]MessageHandler)

//MessageHandler requires the Handle method which handles incoming websocket RPC messages.
//If successful, the bytes returned should be the result of json.Marashal
type MessageHandler interface {
	Handle(userID int) ([]byte, error)
}

func registerHandler(name string, handler MessageHandler) {
	handlers[name] = handler
}

func initWeb() {
	http.HandleFunc("/", index)
	http.HandleFunc("/ws", initWebsocket)

	registerHandler("Tickets", Tickets{})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
	<head>
	<body>
	<script>
		const ws = new WebSocket('ws://localhost:8080/ws');
		ws.onmessage = e => {
			// const message = JSON.parse(e.data);
			// console.log(message.Result)
		}
		ws.onopen = () => {
			for(let i = 0; i < 1000; i++) {
				ws.send(JSON.stringify({ID: i, Method: "Hello", Params: {Name: "mort"}}))
			}
			// ws.send(JSON.stringify({ID: 2, Method: "Fail", Params: {Message: "fail whale"}}))
			// ws.send(JSON.stringify({ID: 3, Method: "sayHello", Params: {Name: "streats"}}))
		}
		document.write("Hello")
	</script>
	</body>
	</html>`))
}

func initWebsocket(w http.ResponseWriter, r *http.Request) {
	//Check user is authenticated
	token := r.Header.Get("API-Key")
	user, err := validateToken([]byte(token))
	if err != nil {
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
	var res WSResp
	var resErr error

	// method string, params json.RawMessage
	if t, ok := handlers[req.Method]; ok {
		// fmt.Printf("Running %s\n", req.Method)
		//Create new copy of the struct and fill its values
		handler := reflect.New(reflect.TypeOf(t))

		handlerInterface := handler.Interface()
		json.Unmarshal(req.Params, &handlerInterface)

		//Call method "Handle" which is garenteed to be avialable due to the interface
		handle, _ := reflect.TypeOf(handlerInterface).MethodByName("Handle")
		returnValues := handle.Func.Call([]reflect.Value{handler, reflect.ValueOf(userID)})

		//Get method return values and handle error caveats
		res.Result = returnValues[0].Interface().([]byte)
		if !returnValues[1].IsNil() {
			resErr = returnValues[1].Elem().Interface().(error)
		}
	} else {
		resErr = fmt.Errorf("Handler %s not found", req.Method)
	}

	res.ID = req.ID

	if resErr != nil {
		tmp := resErr.Error()
		res.Error = &tmp
	}

	responseChan <- res
}
