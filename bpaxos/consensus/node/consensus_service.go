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
		message_stub := common.MessageEvent{&v_stub, []byte("Hello"), []*common.Vertex{&v_stub}}
		return message_stub
	} else {
		return common.MessageEvent{}
	}
}

func (consensusServiceNode *ConsensusServiceNode) ReachConsensus() bool {
	return true
}

func ProcessConsensusMessage(m *nats.Msg, nc *nats.Conn, ctx context.Context, cons *ConsensusServiceNode) {
	fmt.Println("Received proposer to consensus")
	data := common.MessageEvent{}
	json.Unmarshal(m.Data, &data)
	newMessage := cons.HandleReceive(&data)
	sentMessage, err := json.Marshal(&newMessage)

	if err == nil {
		fmt.Println("consensus can publish a message to proposer")
		messenger.PublishNatsMessage(ctx, nc, common.CONSENSUS_TO_PROPOSER, sentMessage)

	} else {
		fmt.Println("json marshal failed")
		fmt.Println(err.Error())
	}
}

func StartConsensus(ctx context.Context, nc *nats.Conn) {
	cons := ConsensusServiceNode{}

	go func(nc *nats.Conn, cons *ConsensusServiceNode) {

		NatsMessage := make(chan *nats.Msg)

		err := messenger.SubscribeToInbox(ctx, nc, common.PROPOSER_TO_CONSENSUS, NatsMessage)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe PROPOSER_TO_CONSENSUS")
		}

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-NatsMessage:
				ProcessConsensusMessage(natsMsg, nc, ctx, cons)
			}
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
