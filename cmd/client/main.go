package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	pb "github.com/jjg-akers/socketses/internal/proto"
	"google.golang.org/protobuf/proto"

	h "github.com/jjg-akers/socketses/cmd/clientlib/handlers"
	s "github.com/jjg-akers/socketses/cmd/clientlib/socket"
	t "github.com/jjg-akers/socketses/internal/types"
)

// use global var for address
var (
	addr    = flag.String("addr", "localhost:8002", "http service addres")
	srvAddr = flag.String("srvAddr", ":8003", "http server address")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

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
	permChan := make(chan t.Message, 10)

	go func() {
		fmt.Println("starting per chan listener")

		// listen for permission requests
		for msg := range permChan {

			protoMsg := pb.Message{
				Type: msg.Type,
				Key:  msg.Key,
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
