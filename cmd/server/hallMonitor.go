package main

import (
	"fmt"
	"log"
)

//This will use a slice to maintain maintain chronological order

//PermissionMsg will hold a key of the device_id-identity_name and
// a channel to signal its ok to proceed
type PermissionMsg struct {
	Key string
	Id  int
}

type GateKeeper struct {
	DoneCh     chan PermissionMsg
	OkCh       chan PermissionMsg
	Permission chan PermissionMsg
}

func NewKeeper() *GateKeeper {
	return &GateKeeper{
		DoneCh:     make(chan PermissionMsg, 10),
		OkCh:       make(chan PermissionMsg, 10),
		Permission: make(chan PermissionMsg, 10),
	}
}

func (g *GateKeeper) Start() {
	log.Println("starting hall monitor")

	register := make(map[string][]int)

	//start listing to channels
	for {
		select {
		case permMsg := <-g.Permission:
			fmt.Println("gk got message on perm channel")
			// check for key
			permSlice, ok := register[permMsg.Key]
			if ok {
				fmt.Println("key in use")

				// some one is using the key
				// add the chan
				register[permMsg.Key] = append(permSlice, permMsg.Id)

			} else {
				fmt.Println("gk else")

				// its not in use,
				// register the key and send ok msg
				register[permMsg.Key] = []int{}

				g.OkCh <- permMsg
			}

		case doneMsg := <-g.DoneCh:
			fmt.Println("gk got message on done channel")

			// check the register
			if permSlice, ok := register[doneMsg.Key]; ok {
				log.Println("found key in register")
				// pop the first item (the oldest) off the slice
				if len(permSlice) > 0 {

					nextUp := permSlice[0]
					register[doneMsg.Key] = permSlice[1:]

					// send the next up to broadcast
					g.OkCh <- PermissionMsg{
						Key: doneMsg.Key,
						Id:  nextUp,
					}
					continue
				}

				log.Println("deleting key: ", doneMsg.Key)

				delete(register, doneMsg.Key)
			}
		} // end select

	} //end for

}
