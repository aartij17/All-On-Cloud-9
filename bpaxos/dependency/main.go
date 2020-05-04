package main

import (
	"All-On-Cloud-9/bpaxos/dependency/node"
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
)

func main() {
	dependency_node := depsnode.NewDepsServiceNode()
	// dependency_node.Stub()
	go func(dep_node *depsnode.DepsServiceNode) {
		socket := common.Socket{}
		_ = socket.Connect(nats.DefaultURL)
		socket.Subscribe(common.LeaderToDeps, func(m *nats.Msg) {
			fmt.Println("Received leader to deps")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			newMessage := dep_node.HandleReceive(&data)
			sentMessage, err := json.Marshal(&newMessage)

			if err == nil {
				fmt.Println("deps can publish a message to leader")
				socket.Publish(common.DepsToLeader, sentMessage)
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}

		})
	}(&dependency_node)
	common.HandleInterrupt()
}
