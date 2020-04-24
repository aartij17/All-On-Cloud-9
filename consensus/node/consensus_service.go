package consensus

import (
	"All-On-Cloud-9/common"
	"fmt"
)

type ConsensusServiceNode struct {
	
}

func (consensusServiceNode *ConsensusServiceNode) HandleReceive(message *common.MessageEvent) {
	if consensusServiceNode.ReachConsensus() {
		v_stub := common.Vertex{0,0}
		message_stub := common.MessageEvent{&v_stub, "Hello", []*common.Vertex{&v_stub}}
		consensusServiceNode.SendResult(&message_stub)
	}
}

func (consensusServiceNode *ConsensusServiceNode) ReachConsensus() bool {
	return true
}

func (consensusServiceNode *ConsensusServiceNode) SendResult(message *common.MessageEvent) {
	fmt.Println("send consensus result")
}