package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/websocket"
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

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Println("Connecting to: ", u.String())

	header := http.Header{}
	header.Set("id", "1")
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("dial failed ", err)
	}
	defer c.Close()

	// set up chan
	done := make(chan struct{})
	okChan := make(chan string, 10)

	// get a SocketMaster
	sm := &s.SocketMaster{
		Conn:     c,
		DoneChan: done,
		OkChan:   okChan,
	}

	// starte a func to listen for broadcast messages
	sm.Start()

	// start a routine for handling requests
	permChan := make(chan t.Message, 10)

	go func() {

		// listen for permission requests
		for msg := range permChan {

			j, err := json.Marshal(msg)
			if err != nil {
				fmt.Println("err marsahlling: ", err)
			}
			fmt.Println("got message on perm chan: ", string(j))

			//send to websocket
			//c.WriteMessage(websocket.TextMessage, []byte(msg))
			sm.Conn.WriteJSON(msg)

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
		if err := server.ListenAndServe(); err != nil {
			fmt.Println("Error in HTTP server:", err)
		}
	}()
	<-i
	wg.Done()
}
