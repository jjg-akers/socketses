package socket

import (
	"log"

	"github.com/gorilla/websocket"
)

type SocketMaster struct {
	Conn     *websocket.Conn
	DoneChan chan struct{}
	OkChan   chan string
}

func (s *SocketMaster) Start() {
	go s.Listen()
}

// starts a func to listen for broadcast messages
func (s *SocketMaster) Listen() {
	defer close(s.DoneChan)
	for {
		_, message, err := s.Conn.ReadMessage()
		if err != nil {
			log.Println("error reading message: ", err)
			return
		}
		log.Println("Message received: ", string(message))

		s.OkChan <- string(message)
	}
}

// func (s SocketMaster)
