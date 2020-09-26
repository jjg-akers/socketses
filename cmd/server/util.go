package main

import (
	"fmt"
	"time"
)

func doWork() {
	fmt.Println("doing work")

	time.Sleep(time.Second * 2)

	fmt.Println("done with work")
}
