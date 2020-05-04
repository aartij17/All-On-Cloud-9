package main

import (
	"All-On-Cloud-9/bpaxos/proposer/node"
	"All-On-Cloud-9/common"
	"fmt"
	"github.com/nats-io/nats.go"
)

func main() {
	socket := common.Socket{}
	_ = socket.Connect(nats.DefaultURL)
	p := proposer.Proposer{}

	go func(proposer *proposer.Proposer, socket *common.Socket) {
		socket.Subscribe(common.LeaderToProposer, func(m *nats.Msg) {
			fmt.Println("Received leader to proposer")
			socket.Publish(common.ProposerToConsensus, m.Data)
		})
	}(&p, &socket)

	go func(proposer *proposer.Proposer, socket *common.Socket) {
		socket.Subscribe(common.ConsensusToProposer, func(m *nats.Msg) {
			fmt.Println("Received consensus to proposer")
			socket.Publish(common.ProposerToReplica, m.Data)
		})
	}(&p, &socket)
	common.HandleInterrupt()
}
