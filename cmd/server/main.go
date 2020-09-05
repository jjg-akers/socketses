package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// using globals for simplicity
var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
	upgrader  = websocket.Upgrader{} // used to upgrade normal http connections
)

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `jseon:"message"`
}

// HandleMessages reads messages from the broadcast channel and passes to clients
func handleMessages() {
	for {
		// wait for next message
		msg := <-broadcast

		// now broadcast that shit!
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Println("client faile to write json", err)

				// for now, if there is an error writing to the client just closs the connection and remove them
				client.Close()
				delete(clients, client)
			}
		}
	}

}

// HandleConns will manage incoming websocket connections
func handleConns(w http.ResponseWriter, r *http.Request) {

	// upgrade the request
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("upgrader failed ", err)
	}

	defer ws.Close()

	//Add to ou glocal clients map
	clients[ws] = true

	// now wait for messages
	for {
		var msg Message

		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading json ", err)
			break
		}

		_ := json.Unmarshal(msg)
		log.Println("message recevied: ", json.Unmarshal(msg))

		// send the message to the broadcast chan
		broadcast <- msg
	}

}

func main() {

	// setup websocket route
	http.HandleFunc("/ws", handleConns)

	// start our message handler routine
	go handleMessages()

	// start servin!
	log.Println("START YOUR SERVEEEERRRRRS")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("Listen and serve failed: ", err)
	}

}
