package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"

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

type Hello struct {
	Name string
}

func (h Hello) Handle(userID int) ([]byte, error) {
	return json.Marshal(fmt.Sprintf("Hello %s with ID of %d", h.Name, userID))
}

type Fail struct {
	Message string
}

func (h Fail) Handle(userID int) ([]byte, error) {
	return nil, errors.New(h.Message)
}

func registerHandler(name string, handler MessageHandler) {
	handlers[name] = handler
}

func initWeb() {
	http.HandleFunc("/", index)
	http.HandleFunc("/ws", initWebsocket)

	registerHandler("Hello", Hello{})
	registerHandler("Fail", Fail{})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
	<head>
	<body>
	<script>
		const ws = new WebSocket('ws://localhost:8080/ws');
		ws.onmessage = e => {
			const message = JSON.parse(e.data);
			console.log(message.Result)
		}
		ws.onopen = () => {
			ws.send(JSON.stringify({ID: 1, Method: "Hello", Params: {Name: "mort"}}))
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
	//Loop all incoming messages
	for {
		_, reader, err := conn.NextReader()
		if err != nil {
			fmt.Printf("Error getting next WS reader, closing connection: %s\n", err.Error())
			conn.Close()
			return
		}

		var req WSReq
		dec := json.NewDecoder(reader)
		dec.Decode(&req)

		//Handle request
		res, respErr := handleRequest(userID, req.Method, req.Params)

		var resp WSResp
		resp.ID = req.ID
		if respErr != nil {
			tmp := respErr.Error()
			resp.Error = &tmp
		} else {
			resp.Result = res
		}

		//Write response message
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			fmt.Printf("Error getting next WS writer, closing connection: %s\n", err.Error())
			conn.Close()
			return
		}

		enc := json.NewEncoder(w)
		err = enc.Encode(resp)
		if err != nil {
			fmt.Printf("Error encoding response: %s\n", err.Error())
		}

		err = w.Close()
		if err != nil {
			fmt.Printf("Error closing ws writer %s\n", err.Error())
		}
	}
}

func handleRequest(userID int, method string, params json.RawMessage) ([]byte, error) {
	if t, ok := handlers[method]; ok {
		//Create new copy of the struct and fill its values
		handler := reflect.New(reflect.TypeOf(t))

		handlerInterface := handler.Interface()
		json.Unmarshal(params, &handlerInterface)

		//Call method "Handle" which is garenteed to be avialable due to the interface
		handle, _ := reflect.TypeOf(handlerInterface).MethodByName("Handle")
		returnValues := handle.Func.Call([]reflect.Value{handler, reflect.ValueOf(userID)})

		//Get method return values and handle error caveats
		res := returnValues[0].Interface().([]byte)
		var err error
		if !returnValues[1].IsNil() {
			err = returnValues[1].Elem().Interface().(error)
		}

		return res, err
	}
	return nil, fmt.Errorf("Handler %s not found", method)
}
