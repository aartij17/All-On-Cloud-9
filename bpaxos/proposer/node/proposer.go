package proposer

import (
	"All-On-Cloud-9/common"
	"fmt"
	"github.com/nats-io/nats.go"
)

type Proposer struct {
}

func (proposer *Proposer) HandleReceive(message *common.MessageEvent) {
	proposer.SendResult(message)
}

func (proposer *Proposer) SendResult(message *common.MessageEvent) {
	fmt.Println("send consensus result")
}

func StartProposer() {
	socket := common.Socket{}
	_ = socket.Connect(nats.DefaultURL)
	p := Proposer{}

	go func(proposer *Proposer, socket *common.Socket) {
		socket.Subscribe(common.LeaderToProposer, func(m *nats.Msg) {
			fmt.Println("Received leader to proposer")
			socket.Publish(common.ProposerToConsensus, m.Data)
		})
	}(&p, &socket)

	go func(proposer *Proposer, socket *common.Socket) {
		socket.Subscribe(common.ConsensusToProposer, func(m *nats.Msg) {
			fmt.Println("Received consensus to proposer")
			socket.Publish(common.ProposerToReplica, m.Data)
		})
	}(&p, &socket)
	common.HandleInterrupt()
}