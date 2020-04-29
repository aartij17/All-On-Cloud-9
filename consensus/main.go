package main

import (
	"All-On-Cloud-9/consensus/node"
	"All-On-Cloud-9/common"
	"github.com/nats-io/nats.go"
	"fmt"
	"encoding/json"
)

func main() {
	cons := consensus.ConsensusServiceNode{}
	go func(cons *consensus.ConsensusServiceNode) {
		socket := common.Socket{}
		_ = socket.Connect(nats.DefaultURL)
		socket.Subscribe(common.ProposerToConsensus, func(m *nats.Msg) {
			fmt.Println("Received proposer to consensus")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			newMessage := cons.HandleReceive(&data)
			sentMessage, err := json.Marshal(&newMessage)

			if err == nil {
				fmt.Println("consensus can publish a message to proposer")
				socket.Publish(common.ConsensusToProposer, sentMessage)
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}

		})
	}(&cons)
	common.HandleInterrupt()
}