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

func StartProposer(ctx context.Context) {
	nc, err := messenger.NatsConnect(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error Proposer connecting to nats server")
		return
	}
	p := Proposer{}

	go func(nc *nats.Conn, proposer *Proposer) {
		_, err = nc.Subscribe(common.LeaderToProposer, func(m *nats.Msg) {
			fmt.Println("Received leader to proposer")
			err = nc.Publish(common.ProposerToConsensus, m.Data)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error publish ProposerToConsensus")
			}
		})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe LeaderToProposer")
		}
	}(nc, &p)

	go func(nc *nats.Conn, proposer *Proposer) {
		_, err = nc.Subscribe(common.ConsensusToProposer, func(m *nats.Msg) {
			fmt.Println("Received consensus to proposer")
			err = nc.Publish(common.ProposerToReplica, m.Data)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("error publish ProposerToReplica")
			}
		})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe ConsensusToProposer")
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
