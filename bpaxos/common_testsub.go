package main

import (
	"fmt"
	"All-On-Cloud-9/common"
	"github.com/nats-io/nats.go"
	"time"
)

func main() {
	// STUB FILE REMOVE LATER
	socket := common.Socket{}
	_ = socket.Connect(nats.DefaultURL)
	socket.Subscribe(common.LeaderToDeps, func(m *nats.Msg) {
		fmt.Println("Received")
	})
	time.Sleep(10000 * time.Millisecond)
}