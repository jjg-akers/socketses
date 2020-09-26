package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	t "github.com/jjg-akers/socketses/internal/types"
)

type ConnHandler struct {
	Permission chan PermissionMsg
	DoneChan   chan PermissionMsg
}

func (c *ConnHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("startedf websocket conn handler")
	// get client id
	id, err := strconv.Atoi(r.Header.Get("id"))
	if err != nil {
		log.Fatal("id conversion failed ", err)

	}

	fmt.Println("got client id: ", id)

	// upgrade the request
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("upgrader failed ", err)
	}

	defer ws.Close()

	// //Add to ou glocal clients map
	// clients[ws] = true

	// register the client
	register[id] = ws

	// now wait for messages
	for {
		var msg t.Message
		//var msg PermissionMsg

		err := ws.ReadJSON(&msg)
		//_, msgBytes, err := ws.ReadMessage()

		if err != nil {
			log.Println("Error reading socket json ", err)
			break
		}

		// check type of message
		if msg.Type == "p" {
			c.Permission <- PermissionMsg{
				Key: msg.Key,
				Id:  id,
			}
			continue
		}

		c.DoneChan <- PermissionMsg{
			Key: msg.Key,
			Id:  id,
		}

		//fmt.Println("got message: ", string(msgBytes))

		// var j []byte
		// err = json.Unmarshal(j, msg)
		// if err != nil {
		// 	log.Println("Error reading json ", err)
		// 	break
		// }
		//og.Println("message recevied: ", string(j))

		// check if key in use
		// 1. register the the key with the hall monitor
		// 2. the hall monitor will be responsible for figureing out who goes next
		// 3. hall monicor sends ok messagage to broadcast channel

		// c.Permission <- PermissionMsg{
		// 	Key: string(msgBytes),
		// 	Id:  id,
		// }
		// wait for response

		// send ok to broadcast
		// doWork()

		// // send the message to the broadcast chan
		// broadcast <- string(msgBytes)
	}

}
