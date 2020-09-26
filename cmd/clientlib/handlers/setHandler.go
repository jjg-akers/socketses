package handlers

import (
	"fmt"
	"net/http"
	"time"

	t "github.com/jjg-akers/socketses/internal/types"
)

type SetHandler struct {
	PermCh chan t.Message
	OkChan chan string
}

func (c *SetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// ask for permission
	fmt.Println("start set handler")
	//time.Sleep(time.Second * 1)
	key := "device_id-ident_key"
	c.PermCh <- t.Message{
		Type: "p",
		Key:  key,
	}

	// fmt.Println("waiting for ok")
	ok := <-c.OkChan
	fmt.Println("got ok: ", ok)

	// fmt.Println("doing some work")
	//time.Sleep(time.Second * 5)

	fmt.Println("done with work")
	c.PermCh <- t.Message{
		Type: "d",
		Key:  key,
	}

	fmt.Println("done time: ", time.Since(start))
	fmt.Fprintln(w, "GOT IT")
}
