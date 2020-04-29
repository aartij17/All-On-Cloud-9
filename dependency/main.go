package main

import (
	"All-On-Cloud-9/dependency/node"
)

func main() {
	dependency_node := depsnode.DepsServiceNode{}
	// dependency_node.Stub()
	go func(dep_node *depsnode.DepsServiceNode) {
		socket := common.Socket{}
		_ = socket.Connect(nats.DefaultURL)
		socket.Subscribe(common.LeaderToDeps, func(m *nats.Msg) {
			fmt.Println("Received")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			newMessage := dep_node.HandleReceive(&data)
			sentMessage, err := json.Marshal(&newMessage)

			if err == nil {
				fmt.Println("i can publish a message")
				socket.Publish(common.DepsToLeader, sentMessage)
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}

		})
	}(&dependency_node)
	common.HandleInterrupt()
}