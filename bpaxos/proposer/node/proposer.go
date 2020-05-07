package proposer

import (
	"All-On-Cloud-9/common"
	"fmt"
	"os"
	"github.com/nats-io/nats.go"
	"All-On-Cloud-9/messenger"
	log "github.com/Sirupsen/logrus"
	"context"
	"os/signal"
)

type Proposer struct {
}

func (proposer *Proposer) HandleReceive(message *common.MessageEvent) {
	proposer.SendResult(message)
}

func (proposer *Proposer) SendResult(message *common.MessageEvent) {
	fmt.Println("send consensus result")
}

func ProcessMessageFromLeader(m *nats.Msg, nc *nats.Conn, ctx context.Context) {
	fmt.Println("Received leader to proposer")
	messenger.PublishNatsMessage(ctx, nc, common.PROPOSER_TO_CONSENSUS, m.Data)
}

func ProcessMessageFromConsensus(m *nats.Msg, nc *nats.Conn, ctx context.Context) {
	fmt.Println("Received consensus to proposer")
	messenger.PublishNatsMessage(ctx, nc, common.PROPOSER_TO_REPLICA, m.Data)
}

func StartProposer(ctx context.Context, nc *nats.Conn) {
	p := Proposer{}

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
				ProcessMessageFromLeader(natsMsg, nc, ctx)
			}
		}
	}(nc, &p)

	go func(nc *nats.Conn, proposer *Proposer) {
		NatsMessage := make(chan *nats.Msg)
		err := messenger.SubscribeToInbox(ctx, nc, common.CONSENSUS_TO_PROPOSER, NatsMessage)

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
				ProcessMessageFromConsensus(natsMsg, nc, ctx)
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
