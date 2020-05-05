package consensus

import (
	"All-On-Cloud-9/common"
	"encoding/json"
	"fmt"
	"context"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	"All-On-Cloud-9/messenger"
	log "github.com/Sirupsen/logrus"
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

func StartConsensus(ctx context.Context) {
	cons := ConsensusServiceNode{}
	nc, err := messenger.NatsConnect(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error Consensus connecting to nats server")
		return
	}
	go func(nc *nats.Conn, cons *ConsensusServiceNode) {
		_, err = nc.Subscribe(common.ProposerToConsensus, func(m *nats.Msg) {
			fmt.Println("Received proposer to consensus")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			newMessage := cons.HandleReceive(&data)
			sentMessage, err := json.Marshal(&newMessage)

			if err == nil {
				fmt.Println("consensus can publish a message to proposer")
				err = nc.Publish(common.ConsensusToProposer, sentMessage)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("error publish ConsensusToProposer")
				}

			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}

		})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe ProposerToConsensus")
		}
	}(nc, &cons)
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
