package proposer

import (
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
	"os"
	"github.com/nats-io/nats.go"
	"All-On-Cloud-9/messenger"
	log "github.com/Sirupsen/logrus"
	"context"
	"os/signal"
)

var (
	id_count = 0
)

type Proposer struct {
	VoteCount int
	Message common.MessageEvent
	ProposerId int
}

func NewProposer() Proposer {
	proposer := Proposer{}
	proposer.VoteCount = 0
	proposer.ProposerId = id_count
	id_count += 1
	return proposer
}

func (proposer *Proposer) HandleReceive(message *common.MessageEvent) {
	proposer.SendResult(message)
}

func (proposer *Proposer) SendResult(message *common.MessageEvent) {
	fmt.Println("send consensus result")
}

func (proposer *Proposer) ProcessMessageFromLeader(m *nats.Msg, nc *nats.Conn, ctx context.Context) {
	fmt.Println("Received leader to proposer")
	data := common.MessageEvent{}
	err := json.Unmarshal(m.Data, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err":            err.Error(),
		}).Error("error unmarshal message from leader")
		return
	}
	proposer.VoteCount = 0
	proposer.Message = data
	vertexId := data.VertexId
	// TODO: hardcoded the ip address for now, change that later
	consensusMessage := common.ConsensusMessage{VertexId:vertexId, ProposerId:proposer.ProposerId, Release: 0}
	sentMessage, err := json.Marshal(&consensusMessage)
	if err != nil {
		log.WithFields(log.Fields{
			"err":            err.Error(),
		}).Error("error marshal consensus message")
		return
	}
	messenger.PublishNatsMessage(ctx, nc, common.PROPOSER_TO_CONSENSUS, sentMessage)
}

func (proposer *Proposer) ProcessMessageFromConsensus(m *nats.Msg, nc *nats.Conn, ctx context.Context) {
	fmt.Println("Received consensus to proposer")
	data := common.Vertex{}
	err := json.Unmarshal(m.Data, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err":            err.Error(),
		}).Error("error unmarshal message from leader")
		return
	}

	if (data.Index == proposer.Message.VertexId.Index) && (data.Id == proposer.Message.VertexId.Id) {
		proposer.VoteCount += 1
	}

	if proposer.VoteCount > common.F {
		replicaMessage, err := json.Marshal(&proposer.Message)
		if err != nil {
			log.WithFields(log.Fields{
				"err":            err.Error(),
			}).Error("error marshal proposer message")
			return
		}
		messenger.PublishNatsMessage(ctx, nc, common.PROPOSER_TO_REPLICA, replicaMessage)
		consensusMessage := common.ConsensusMessage{VertexId:proposer.Message.VertexId, ProposerId:proposer.ProposerId, Release: 1}
		sentConsensusMessage, err := json.Marshal(&consensusMessage)
		if err != nil {
			log.WithFields(log.Fields{
				"err":            err.Error(),
			}).Error("error marshal proposer release message")
			return
		}
		messenger.PublishNatsMessage(ctx, nc, common.PROPOSER_TO_CONSENSUS, sentConsensusMessage)
	}
}

func StartProposer(ctx context.Context, nc *nats.Conn) {
	p := NewProposer()

	go func(nc *nats.Conn, proposer *Proposer) {
		NatsMessage := make(chan *nats.Msg)
		err := messenger.SubscribeToInbox(ctx, nc, common.LEADER_TO_PROPOSER, NatsMessage)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe LEADER_TO_PROPOSER")
		}

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-NatsMessage:
				proposer.ProcessMessageFromLeader(natsMsg, nc, ctx)
			}
		}
	}(nc, &p)

	go func(nc *nats.Conn, proposer *Proposer) {
		NatsMessage := make(chan *nats.Msg)
		sub := fmt.Sprintf("%s%d", common.CONSENSUS_TO_PROPOSER, proposer.ProposerId)
		err := messenger.SubscribeToInbox(ctx, nc, sub, NatsMessage)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe CONSENSUS_TO_PROPOSER")
		}

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-NatsMessage:
				proposer.ProcessMessageFromConsensus(natsMsg, nc, ctx)
			}
		}
	}(nc, &p)
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			log.Info("Received an interrupt, stopping all connections...")
			//cancel()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
