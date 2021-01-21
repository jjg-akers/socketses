package main

import (
	"bufio"
	"fmt"
	"log"
	"net"

	pb "github.com/jjg-akers/socketses/internal/proto"
	t "github.com/jjg-akers/socketses/internal/types"
	"google.golang.org/protobuf/proto"
)

type ConnHandler struct {
	Permission chan PermissionMsg
	DoneChan   chan DoneMsg
}

func (c *ConnHandler) HandleConn(conn *net.TCPConn) {
	fmt.Println("startd conn handler")
	defer conn.Close()

	// get client id
	//var b = make([]byte, 8)
	// var id int64

	// n, err := conn.Read(b[:])
	// if err != nil {
	// 	log.Println("conn read failed: ", err)
	// }

	// err := binary.Read(conn, binary.LittleEndian, &id)
	// if err != nil {
	// 	log.Println("binary read failed: ", err)
	// }
	// id, err := strconv.Atoi(string(id))
	// if err != nil {
	// 	log.Fatal("id conversion failed ", err)

	// }

	r := bufio.NewReader(conn)
	registrationMsg := pb.Register{}
	//var msg PermissionMsg

	data, err := r.ReadBytes('|')
	if err != nil {
		fmt.Println("error reading bytes form buffer")
		_, err := conn.Write([]byte("false|"))
		if err != nil {
			log.Println("could not write bytes to conn", err)
		}
		return
	}

	data = data[:len(data)-1]
	err = proto.Unmarshal(data, &registrationMsg)
	if err != nil {
		fmt.Println("error reading bytes form buffer")
		_, err := conn.Write([]byte("false|"))
		if err != nil {
			log.Println("could not write bytes to conn", err)
		}
		return
	}

	fmt.Println("got client id: ", registrationMsg.GetId())
	fmt.Println("registering type: ", registrationMsg.GetType())

	// msgBytes, err := proto.Marshal(&protoMsg)
	// if err != nil {
	// 	log.Println("error marshalling proto n", err)
	// 	continue
	// }

	// msgBytes = append(msgBytes, '|')

	// i, err := client.Write(msgBytes)
	// if err != nil {
	// 	log.Println("could not write bytes to conn", err)

	// 	client.Close()
	// 	delete(clients, client)
	// }

	// upgrade the request
	// ws, err := upgrader.Upgrade(w, r, nil)
	// if err != nil {
	// 	log.Fatal("upgrader failed ", err)
	// }

	// //Add to ou glocal clients map
	// clients[ws] = true

	// register the client
	if registrationMsg.GetType() == "permission" {
		register[registrationMsg.GetId()] = conn

		// now wait for messages
		//r := bufio.NewReader(conn)
		for {
			msg := pb.Permission{}
			//var msg PermissionMsg

			data, err := r.ReadBytes('|')
			if err != nil {
				//fmt.Println("error reading bytes form buffer")
				continue
			}

			fmt.Println("proto data: ", string(data))

			data = data[:len(data)-1]
			err = proto.Unmarshal(data, &msg)
			if err != nil {
				fmt.Println("error unmarshalling proto")
				continue
			}

			// err := ws.ReadJSON(&msg)
			//_, msgBytes, err := ws.ReadMessage()

			// if err != nil {
			// 	log.Println("Error reading socket json ", err)
			// 	break
			// }

			// check type of message
			if msg.GetKey().GetType() == "p" {
				c.Permission <- PermissionMsg{
					Key: msg.GetKey().GetKey(),
					Id:  registrationMsg.GetId(),
				}
				continue
			}

			c.DoneChan <- DoneMsg{
				Key: msg.GetKey().GetKey(),
				Id:  registrationMsg.GetId(),
				Update: t.CacheUpdate{
					Key:        msg.GetCacheUpdate().GetKey(),
					Value:      msg.GetCacheUpdate().GetValue(),
					Expiration: msg.GetCacheUpdate().GetExpiration(),
				},
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

	// registration type is not permission
	if registrationMsg.GetType() == "update" {
		// register in update map
		updateRegister[registrationMsg.GetId()] = conn

	}

}
