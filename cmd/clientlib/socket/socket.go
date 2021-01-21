package socket

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/jjg-akers/socketses/internal/proto"
	"google.golang.org/protobuf/proto"
)

type SocketMaster struct {
	PermissionConn *net.TCPConn
	UpdateConn     *net.TCPConn
	DoneChan       chan struct{}
	OkChan         chan string
}

func (s *SocketMaster) Start(id int64, interupt chan os.Signal) {
	fmt.Println("starting socketmastter")

	conn, err := getConn("tcp", ":8002")
	if err != nil {
		log.Fatal("Failed to create tcp addr: ", err)
	}
	// send id before we send other messages
	err = binary.Write(conn, binary.LittleEndian, id)
	//err := binary.Read(conn, binary.LittleEndian, &id)
	if err != nil {
		log.Println("binary Write failed: ", err)
	}

	defer conn.Close()
	s.PermissionConn = conn
	s.Listen(interupt)
}

func getConn(network string, address string) (*net.TCPConn, error) {
	addr, err := net.ResolveTCPAddr("tcp", ":8002")
	if err != nil {
		return nil, err
		// log.Fatal("Failed to create tcp addr: ", err)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	//c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		// log.Fatal("dial failed ", err)
		return nil, err
	}

	return conn, nil

}

// starts a func to listen for broadcast messages
func (s *SocketMaster) Listen(interupt chan os.Signal) {
	fmt.Println("starting socketmastter listen")

	defer close(s.DoneChan)

	r := bufio.NewReader(s.Conn)
	for {
		select {
		case <-interupt:
			fmt.Println("exiting")
			return
		default:

			msg := pb.Permission{}
			//var msg PermissionMsg

			data, err := r.ReadBytes('|')
			if err != nil {
				fmt.Println("error reading bytes form buffer")
				continue
			}

			data = data[:len(data)-1]
			err = proto.Unmarshal(data, &msg)
			if err != nil {
				fmt.Println("error unmarshalling proto")
				continue
			}

			fmt.Println("got message: ", msg.GetKey())

			s.OkChan <- string(msg.GetKey().Key)
		}
	}
}

// defer close(s.DoneChan)
// for {
// 	_, message, err := s.Conn.ReadMessage()
// 	if err != nil {
// 		log.Println("error reading message: ", err)
// 		return
// 	}
// 	log.Println("Message received: ", string(message))

// 	s.OkChan <- string(message)
// }

// func (s *SocketMaster) HandleConn(conn *net.TCPConn) {
// 	r := bufio.NewReader(conn)
// 	for {
// 		msg := pb.Message{}
// 		//var msg PermissionMsg

// 		data, err := r.ReadBytes('|')
// 		if err != nil {
// 			fmt.Println("error reading bytes form buffer")
// 			continue
// 		}

// 		err = proto.Unmarshal(data, &msg)
// 		if err != nil {
// 			fmt.Println("error unmarshalling proto")
// 			continue
// 		}

// }
