package consensus

import (
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
)

type ConsensusServiceNode struct {
}

func (consensusServiceNode *ConsensusServiceNode) HandleReceive(message *common.MessageEvent) common.MessageEvent {
	if consensusServiceNode.ReachConsensus() {
		v_stub := common.Vertex{0, 0}
		message_stub := common.MessageEvent{&v_stub, "Hello", []*common.Vertex{&v_stub}}
		return message_stub
	} else {
		return common.MessageEvent{}
	}
}

func (consensusServiceNode *ConsensusServiceNode) ReachConsensus() bool {
	return true
}

func StartConsensus() {
	cons := ConsensusServiceNode{}
	go func(cons *ConsensusServiceNode) {
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