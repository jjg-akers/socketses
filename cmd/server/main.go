package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// using globals for simplicity
var (
	clients = make(map[*websocket.Conn]bool)
	// broadcast = make(chan Message)
	broadcast = make(chan string)
	upgrader  = websocket.Upgrader{} // used to upgrade normal http connections
)

// type Message struct {
// 	Type string `json:"type"`
// 	Key  string `json:"key"`
// }

var register = make(map[int]*websocket.Conn)

var que = make(map[string][]int)

// HandleMessages reads messages from the broadcast channel and passes to clients
func handleMessages(okChan chan PermissionMsg) {
	for {
		// 1. wait for message
		// 2. get the conn for the client id
		// 3. send ok

		// wait for next message
		msg := <-okChan

		fmt.Println("got message on broadcast: ", msg)

		// now broadcast that shit!
		if client, ok := register[msg.Id]; ok {
			err := client.WriteJSON(msg)

			//err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("client faile to write json", err)

				// for now, if there is an error writing to the client just closs the connection and remove them
				client.Close()
				delete(clients, client)
			}
		}

		// for client := range clients {
		// 	//err := client.WriteJSON(msg)
		// 	err := client.WriteMessage(websocket.TextMessage, []byte(msg))
		// 	if err != nil {
		// 		log.Println("client faile to write json", err)

		// 		// for now, if there is an error writing to the client just closs the connection and remove them
		// 		client.Close()
		// 		delete(clients, client)
		// 	}
		// }
	}

}

// HandleConns will manage incoming websocket connections
func handleConns(w http.ResponseWriter, r *http.Request) {

}

var hm = NewKeeper()

func main() {

	go hm.Start()
	// start the hall monitor

	// setup websocket route
	//http.HandleFunc("/ws", ConnHandler{Permission: hm.Permission})
	http.Handle("/ws", &ConnHandler{
		Permission: hm.Permission,
		DoneChan:   hm.DoneCh,
	})

	// start our message handler routine
	go handleMessages(hm.OkCh)

	// start servin!
	log.Println("START YOUR SERVEEEERRRRRS")
	err := http.ListenAndServe(":8002", nil)
	if err != nil {
		log.Fatal("Listen and serve failed: ", err)
	}

}
