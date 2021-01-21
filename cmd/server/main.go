package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	pb "github.com/jjg-akers/socketses/internal/proto"
	t "github.com/jjg-akers/socketses/internal/types"
	"google.golang.org/protobuf/proto"
)

// using globals for simplicity
var (
	clients = make(map[*net.TCPConn]bool)
	// broadcast = make(chan Message)
	broadcast = make(chan string)
	upgrader  = websocket.Upgrader{} // used to upgrade normal http connections
)

// type Message struct {
// 	Type string `json:"type"`
// 	Key  string `json:"key"`
// }

var register = make(map[int64]*net.TCPConn)
var updateRegister = make(map[int64]*net.TCPConn)

var que = make(map[string][]int64)

// HandleMessages reads messages from the broadcast channel and passes to clients
func handleMessages(okChan chan PermissionMsg) {
	for {
		// 1. wait for message
		// 2. get the conn for the client id
		// 3. send ok

		// wait for next message
		msg := <-okChan
		protoMsg := pb.Permission{
			Key: &pb.Key{
				Key:  msg.Key,
				Type: "ok",
			},
		}

		fmt.Println("got message on broadcast: ", msg)

		// now broadcast that shit!
		if client, ok := register[msg.Id]; ok {
			msgBytes, err := proto.Marshal(&protoMsg)
			if err != nil {
				log.Println("error marshalling proto n", err)
				continue
			}

			msgBytes = append(msgBytes, '|')

			i, err := client.Write(msgBytes)
			if err != nil {
				log.Println("could not write bytes to conn", err)

				client.Close()
				delete(clients, client)
			}

			fmt.Println("number of bytes writting: ", i)

			// err := client.WriteJSON(msg)

			// //err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			// if err != nil {
			// 	log.Println("client faile to write json", err)

			// 	// for now, if there is an error writing to the client just closs the connection and remove them
			// 	client.Close()
			// 	delete(clients, client)
			// }
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

func handleUpdateMessages(updateChan chan t.CacheUpdate) {
	for {
		// 1. wait for message
		// 2. get the conn for the client id
		// 3. send ok

		// wait for next message
		msg := <-updateChan
		protoMsg := pb.CacheUpdate{
			Key:        msg.Key,
			Value:      msg.Value,
			Expiration: msg.Expiration,
		}

		fmt.Println("got message on update handler: ", msg)

		// now broadcast that shit!
		for _, client := range updateRegister {
			msgBytes, err := proto.Marshal(&protoMsg)
			if err != nil {
				log.Println("error marshalling proto n", err)
				continue
			}

			msgBytes = append(msgBytes, '|')

			i, err := client.Write(msgBytes)
			if err != nil {
				log.Println("could not write bytes to conn", err)

				client.Close()
				delete(clients, client)
			}

			fmt.Println("number of bytes writting: ", i)

			// err := client.WriteJSON(msg)

			// //err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			// if err != nil {
			// 	log.Println("client faile to write json", err)

			// 	// for now, if there is an error writing to the client just closs the connection and remove them
			// 	client.Close()
			// 	delete(clients, client)
			// }
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
	// http.Handle("/ws", &ConnHandler{
	// 	Permission: hm.Permission,
	// 	DoneChan:   hm.DoneCh,
	// })

	wg := sync.WaitGroup{}
	wg.Add(1)
	go startServer(&wg, hm.Permission, hm.DoneCh)
	// start our message handler routine
	go handleMessages(hm.OkCh)
	wg.Wait()

}

func startServer(wg *sync.WaitGroup, p chan PermissionMsg, d chan DoneMsg) {

	addr, err := net.ResolveTCPAddr("tcp", ":8002")
	if err != nil {
		log.Fatal("Failed to create tcp addr: ", err)
	}

	fmt.Println("starting server")
	li, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("Listen and serve failed: ", err)
	}

	for {
		if conn, err := li.AcceptTCP(); err == nil {
			ch := ConnHandler{
				Permission: p,
				DoneChan:   d,
			}

			go ch.HandleConn(conn)
		}
	}

	wg.Done()

}
