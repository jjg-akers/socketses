package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// use global var for address
var addr = flag.String("addr", "localhost:8000", "http service addres")

type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `jseon:"message"`
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Println("Connecting to: ", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial failed ", err)
	}
	defer c.Close()

	// set up chan
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("error reading message: ", err)
				return
			}
			log.Println("Message received: ", string(message))
		}
	}()

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	m := Message{
		Email:    "something@something.com",
		Username: "hot hot hot",
		Message:  "hellow from client",
	}

	js, _ := json.Marshal(m)

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, js)
			if err != nil {
				log.Println("error client writing message: ", err)
				return
			}
			log.Println("t: ", t.String())

		case <-interrupt:

			log.Println("interupt received: ", err)

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("error closing client: ", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second * 30):
			}
			return

		}
	}
}
