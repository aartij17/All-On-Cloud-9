package consensus

import (
	"All-On-Cloud-9/common"
)

type ConsensusServiceNode struct {
	
}

func (consensusServiceNode *ConsensusServiceNode) HandleReceive(message *common.MessageEvent) common.MessageEvent{
	if consensusServiceNode.ReachConsensus() {
		v_stub := common.Vertex{0,0}
		message_stub := common.MessageEvent{&v_stub, "Hello", []*common.Vertex{&v_stub}}
		return message_stub
	} else {
		return common.MessageEvent{}
	}
}

func (consensusServiceNode *ConsensusServiceNode) ReachConsensus() bool {
	return true
}
