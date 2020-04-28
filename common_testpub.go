package main

import (
	"fmt"
	"All-On-Cloud-9/common"
	"github.com/nats-io/nats.go"
	"encoding/json"
)

func main() {
	// STUB FILE REMOVE LATER
	v0 := common.Vertex{0,0}
	message1 := common.MessageEvent{&v0, "Hello", []*common.Vertex{&v0, &v0}}
	sentMessage, err := json.Marshal(&message1)
	if err != nil {
		fmt.Println("no marshal booboo")
	}
	socket := common.Socket{}
	_ = socket.Connect(nats.DefaultURL)
	socket.Publish(common.DepsToLeader, sentMessage)
	fmt.Println("YAY")
}