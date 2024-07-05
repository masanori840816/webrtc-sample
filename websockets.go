package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	listLock    sync.RWMutex
	connections []connectionState
)

type websocketMessage struct {
	MessageType string `json:"messageType"`
	Data        string `json:"data"`
}
type connectionState struct {
	userName  string
	websocket *threadSafeWriter
}
type threadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getParam(r, "user")
	if err != nil {
		log.Println(err)
		return
	}
	unsafeConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn := &threadSafeWriter{unsafeConn, sync.Mutex{}}
	// Close the connection when the for-loop operation is finished.
	defer conn.Close()
	listLock.Lock()
	connections = append(connections, connectionState{userName: user, websocket: conn})
	listLock.Unlock()

	message := &websocketMessage{}
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err)
			return
		}
		for _, c := range connections {
			if c.userName == user {
				continue
			}
			c.websocket.WriteJSON(message)
		}
	}
}
func (t *threadSafeWriter) WriteJSON(v interface{}) error {
	t.Lock()
	defer t.Unlock()

	return t.Conn.WriteJSON(v)
}
func getParam(r *http.Request, key string) (string, error) {
	result := r.URL.Query().Get(key)
	if len(result) <= 0 {
		return "", fmt.Errorf("no value: %s", key)
	}
	return result, nil
}
