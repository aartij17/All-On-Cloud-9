package main

import (
	"All-On-Cloud-9/bpaxos/leader/node"
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
)

func main() {
	l := leadernode.NewLeader(0)
	socket := common.Socket{}
	_ = socket.Connect(nats.DefaultURL)

	go func(leader *leadernode.Leader, socket *common.Socket) {
		socket.Subscribe(common.DepsToLeader, func(m *nats.Msg) {
			fmt.Println("Received deps to leader")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			l.AddToMessages(&data)
			if l.GetMessagesLen() > common.F {
				newMessageEvent := l.HandleReceiveDeps()
				// fmt.Printf("%T\n", newMessageEvent)
				// fmt.Printf("%T\n", newMessageEvent.VertexId)
				// fmt.Printf("%T\n", newMessageEvent.Deps)
				// fmt.Printf("%+v\n",newMessageEvent)
				sentMessage, err := json.Marshal(&newMessageEvent)
				if err == nil {
					fmt.Println("leader can publish a message to proposer")
					socket.Publish(common.LeaderToProposer, sentMessage)
				} else {
					fmt.Println("json marshal failed")
					fmt.Println(err.Error())
				}
				// should we flush when it fails?
				l.FlushMessages()
			}
		})
	}(&l, &socket)

	go func(leader *leadernode.Leader, socket *common.Socket) {
		socket.Subscribe(common.ClientToLeader, func(m *nats.Msg) {
			fmt.Println("Received client to leader")
			newMessage := leader.HandleReceiveCommand(string(m.Data))
			sentMessage, err := json.Marshal(&newMessage)
			if err == nil {
				fmt.Println("leader can publish a message to deps")
				socket.Publish(common.LeaderToDeps, sentMessage)
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}
		})
	}(&l, &socket)
	common.HandleInterrupt()
}
