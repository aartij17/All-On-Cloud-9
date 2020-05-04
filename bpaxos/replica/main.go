package main

import (
	"All-On-Cloud-9/bpaxos/replica/node"
	"All-On-Cloud-9/common"
	"github.com/nats-io/nats.go"
	"fmt"
	"encoding/json"
)

func main() {
	rep := replica.Replica{}
	go func(rep *replica.Replica) {
		socket := common.Socket{}
		_ = socket.Connect(nats.DefaultURL)
		socket.Subscribe(common.ProposerToReplica, func(m *nats.Msg) {
			fmt.Println("Received proposer to replica")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			newMessage := rep.HandleReceive(&data)
			fmt.Println(newMessage)

		})
	}(&rep)
	common.HandleInterrupt()
}