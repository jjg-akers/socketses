package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	pb "github.com/jjg-akers/socketses/internal/proto"
	"google.golang.org/protobuf/proto"

	h "github.com/jjg-akers/socketses/cmd/clientlib/handlers"
	s "github.com/jjg-akers/socketses/cmd/clientlib/socket"
	t "github.com/jjg-akers/socketses/internal/types"
	"github.com/patrickmn/go-cache"
)

// use global var for address
var (
	addr    = flag.String("addr", "localhost:8002", "http service addres")
	srvAddr = flag.String("srvAddr", ":8003", "http server address")
)

var localCache *cache.Cache

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	localCache = cache.New(time.Minute*2, time.Second*10)

	// u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	// log.Println("Connecting to: ", u.String())

	// header := http.Header{}
	// header.Set("id", "1")

	// set up chan
	done := make(chan struct{})
	okChan := make(chan string, 10)

	// get a SocketMaster
	sm := &s.SocketMaster{
		DoneChan: done,
		OkChan:   okChan,
	}

	var id int64 = 123
	// starte a func to listen for broadcast messages
	go sm.Start(id, interrupt)

	// start a routine for handling requests
	permChan := make(chan t.Permission, 10)

	go func() {
		fmt.Println("starting clinet permision chan listener")

		// listen for permission requests
		for msg := range permChan {

			protoMsg := pb.Permission{
				Key:         msg.Key,
				CacheUpdate: msg.CacheUpdate,
			}

			data, err := proto.Marshal(&protoMsg)
			// j, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("err prorto marsahlling: ", err)
			}

			fmt.Println("sending message on from perm chan: ", protoMsg.String())

			//send to websocket
			//c.WriteMessage(websocket.TextMessage, []byte(msg))
			//sm.Conn.WriteJSON(msg)
			data = append(data, '|')
			i, err := sm.Conn.Write(data)
			if err != nil {
				fmt.Println("err sending data from clinent: ", err)
			}
			fmt.Printf("worte %d bytes to conn\n", i)

		}
	}()

	// start a server
	wg := sync.WaitGroup{}
	wg.Add(1)

	// sh := setHandler{
	// 	permCh: permChan,
	// 	okChan: okChan,
	// }
	go startServer(&wg, interrupt, permChan, okChan)

	wg.Wait()
	// ticker := time.NewTicker(time.Second * 3)
	// defer ticker.Stop()

	// m := Message{
	// 	Email:    "something@something.com",
	// 	Username: "hot hot hot",
	// 	Message:  "hellow from client",
	// }

	// js, _ := json.Marshal(m)

	// for {
	// 	select {
	// 	case <-done:
	// 		return
	// 	case t := <-ticker.C:
	// 		err := c.WriteMessage(websocket.TextMessage, js)
	// 		if err != nil {
	// 			log.Println("error client writing message: ", err)
	// 			return
	// 		}
	// 		log.Println("t: ", t.String())

	// 	case <-interrupt:

	// 		log.Println("interupt received: ", err)

	// 		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 		if err != nil {
	// 			log.Println("error closing client: ", err)
	// 			return
	// 		}
	// 		select {
	// 		case <-done:
	// 		case <-time.After(time.Second * 30):
	// 		}
	// 		return

	// 	}
	// }
}

func startServer(wg *sync.WaitGroup, i chan os.Signal, pCh chan t.Message, ok chan string) {

	mux := http.NewServeMux()

	mux.Handle("/identity/set", &h.SetHandler{
		PermCh: pCh,
		OkChan: ok,
	})

	server := &http.Server{
		Addr:    *srvAddr,
		Handler: mux,
	}

	go func() {
		fmt.Println("starting web server")
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Error in HTTP server:", err)
		}
	}()
	<-i
	wg.Done()
}

//cacheUpdater will receive update messages update the local cache
func cacheUpdater(conn *net.TCPConn) {
	// wait for messages
	r := bufio.NewReader(conn)
	for {
		updateMsg := pb.CacheUpdate{}
		//var msg PermissionMsg

		data, err := r.ReadBytes('|')
		if err != nil {
			fmt.Println("error reading bytes form buffer")
			continue
		}

		fmt.Println("update proto data: ", string(data))

		go func(d []byte) {
			//data = data[:len(data)-1]
			err = proto.Unmarshal(d[:len(data)-1], &updateMsg)
			if err != nil {
				fmt.Println("error unmarshalling proto")
				return
			}

			localCache.Set(updateMsg.GetKey(), updateMsg.GetValue(), 0)
		}(data)

		// err := ws.ReadJSON(&msg)
		//_, msgBytes, err := ws.ReadMessage()

		// if err != nil {
		// 	log.Println("Error reading socket json ", err)
		// 	break
		// }

		// check type of message
		// if msg.Type == "p" {
		// 	c.Permission <- PermissionMsg{
		// 		Key: msg.Key,
		// 		Id:  id,
		// 	}
		// 	continue
		// }

		// c.DoneChan <- PermissionMsg{
		// 	Key: msg.Key,
		// 	Id:  id,
		// }

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
