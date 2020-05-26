package consensus

import (
	"All-On-Cloud-9/bpaxos/debug"
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

var (
	mux          sync.Mutex
	timer        *time.Timer
	timeout_quit = make(chan bool)
)

type ConsensusServiceNode struct {
	VertexId *common.Vertex
}

func NewConsensusServiceNode() ConsensusServiceNode {
	consensusNode := ConsensusServiceNode{}
	consensusNode.VertexId = nil
	return consensusNode
}

func Timeout(consensusNode *ConsensusServiceNode) {
	select {
	case <-timer.C:
		log.Error("gained lock before timeout")
		mux.Lock()
		if consensusNode.VertexId != nil {
			consensusNode.VertexId = nil
			log.Error("Consensus timeout")
		}
		mux.Unlock()
		log.Error("released lock after timeout")
	case <-timeout_quit:
		log.Info("[BPAXOS] consensus Timeout not needed")
	}
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

func (consensusServiceNode *ConsensusServiceNode) ProcessConsensusMessage(m *nats.Msg,
	nc *nats.Conn, ctx context.Context) {
	log.WithFields(log.Fields{
		"consensusNodeId": consensusServiceNode.VertexId,
	}).Info("received message from proposer to consensus")
	data := common.ConsensusMessage{}
	err := json.Unmarshal(m.Data, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("error unmarshal message from proposer")
		return
	}
	if (consensusServiceNode.VertexId == nil) && (data.Release == 0) {
		consensusServiceNode.VertexId = data.VertexId
		sub := fmt.Sprintf("%s%d", common.CONSENSUS_TO_PROPOSER, data.ProposerId)

		sentMessage, err := json.Marshal(&data.VertexId)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("error marshal consensus vertex message")
			return
		}
		messenger.PublishNatsMessage(ctx, nc, sub, sentMessage)
		timer = time.NewTimer(time.Duration(common.CONSENSUS_TIMEOUT_MILLISECONDS) * time.Millisecond)
		go Timeout(consensusServiceNode)
	} else {
		// release vote
		if (consensusServiceNode.VertexId != nil) && (consensusServiceNode.VertexId.Index == data.VertexId.Index) &&
			(consensusServiceNode.VertexId.Id == data.VertexId.Id) && (data.Release == 1) {
			timer.Stop()
			timeout_quit <- true
			consensusServiceNode.VertexId = nil
		}
	}
}

func StartConsensus(ctx context.Context, nc *nats.Conn) {
	cons := ConsensusServiceNode{}

	go func(nc *nats.Conn, cons *ConsensusServiceNode) {

		natsMessage := make(chan *nats.Msg)

		err := messenger.SubscribeToInbox(ctx, nc, common.PROPOSER_TO_CONSENSUS, natsMessage, false)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe PROPOSER_TO_CONSENSUS")
		}
		debug.WriteToFile("C")

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-natsMessage:
				log.Error("gained lock before ProcessConsensusMessage")
				mux.Lock()
				cons.ProcessConsensusMessage(natsMsg, nc, ctx)
				mux.Unlock()
				log.Error("released lock after ProcessConsensusMessage")
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
