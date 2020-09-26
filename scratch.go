package flowcontrol

import (
	"log"
)

//This will use a slice to maintain maintain chronological order

//PermissionMsg will hold a key of the device_id-identity_name and
// a channel to signal its ok to proceed
type PermissionMsg struct {
	Key    string
	OkChan chan bool
}

type GateKeeper struct {
	DoneCh     chan string
	Permission chan PermissionMsg
}

func NewKeeper() *GateKeeper {
	return &GateKeeper{
		DoneCh:     make(chan string),
		Permission: make(chan PermissionMsg),
	}
}

func (g *GateKeeper) Start() {
	log.Println("starting hall monitor")

	register := make(map[string][]chan bool)

	//start listing to channels
	for {
		select {
		case permMsg := <-g.Permission:
			// check for key
			permSlice, ok := register[permMsg.Key]
			if ok {
				// some one is using the key
				// add the chan
				register[permMsg.Key] = append(permSlice, permMsg.OkChan)

			} else {
				// its not in use,
				// register the key and send ok msg
				register[permMsg.Key] = []chan bool{}

				permMsg.OkChan <- true
			}

		case doneMsg := <-g.DoneCh:
			// check the register
			if permSlice, ok := register[doneMsg]; ok {
				// pop the first item (the oldest) off the slice
				if len(permSlice) > 0 {

					nextUp := permSlice[0]
					register[doneMsg] = permSlice[1:]

					nextUp <- true
					continue
				}

				delete(register, doneMsg)
			}
		} // end select

	} //end for

}
