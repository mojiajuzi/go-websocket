package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]bool)

var broadcast = make(chan Message)

// Message 消息结构
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	fs := http.FileServer(http.Dir("/home/gru/go/src/gotemp/websocket/public"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	log.Println("http server started on:9200")
	err := http.ListenAndServe(":9200", nil)
	if err != nil {
		log.Fatal()
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Origin")
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println(err.Error())
			delete(clients, ws)
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println(err.Error())
				client.Close()
				delete(clients, client)
			}
		}
	}
}
