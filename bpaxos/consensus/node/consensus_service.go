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
	VertexId *common.Vertex
}

func NewConsensusServiceNode() ConsensusServiceNode {
	consensusNode := ConsensusServiceNode{}
	consensusNode.VertexId = nil
	return consensusNode
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

func (consensusServiceNode *ConsensusServiceNode) ProcessConsensusMessage(m *nats.Msg, nc *nats.Conn, ctx context.Context, cons *ConsensusServiceNode) {
	fmt.Println("Received proposer to consensus")
	data := common.ConsensusMessage{}
	err := json.Unmarshal(m.Data, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err":            err.Error(),
		}).Error("error unmarshal message from proposer")
		return
	}

	if consensusServiceNode.VertexId == nil {
		consensusServiceNode.VertexId = data.VertexId
		sub := fmt.Sprintf("%s%d", common.CONSENSUS_TO_PROPOSER, data.ProposerId)

		sentMessage, err := json.Marshal(&data.VertexId)
		if err != nil {
			log.WithFields(log.Fields{
				"err":            err.Error(),
			}).Error("error marshal consensus vertex message")
			return
		}
		messenger.PublishNatsMessage(ctx, nc, sub, sentMessage)
	} else {
		// release vote
		if (consensusServiceNode.VertexId.Index == data.VertexId.Index) && (consensusServiceNode.VertexId.Id == data.VertexId.Id) && (data.Release == 1) {
			consensusServiceNode.VertexId = nil
		}
	}
	// newMessage := cons.HandleReceive(&data)
	// sentMessage, err := json.Marshal(&newMessage)

	// if err == nil {
	// 	fmt.Println("consensus can publish a message to proposer")
	// 	messenger.PublishNatsMessage(ctx, nc, common.CONSENSUS_TO_PROPOSER, sentMessage)

	// } else {
	// 	fmt.Println("json marshal failed")
	// 	fmt.Println(err.Error())
	// }
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
				cons.ProcessConsensusMessage(natsMsg, nc, ctx, cons)
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
