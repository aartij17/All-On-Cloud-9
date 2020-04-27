package main

import (
	"fmt"
	"time"
	"All-On-Cloud-9/common"
	"github.com/nats-io/nats.go"
)

func main() {
	// STUB FILE REMOVE LATER
	socket := common.Socket{}
	_ = socket.Connect(nats.DefaultURL)
	socket.Publish(common.LeaderToDeps, []byte("golden retriever"))
	time.Sleep(10000 * time.Millisecond)
	fmt.Println("YAY")
}