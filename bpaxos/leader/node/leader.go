package leadernode

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

var (
	id_count    = 0
	proposer_id = 0 // This ID will determine which proposer to send to
	timer        *time.Timer
	mux          sync.Mutex
)

type Leader struct {
	Index    int
	messages []*common.MessageEvent
	m        map[*common.Vertex]*common.MessageEvent{}
}

func NewLeader(index int) Leader {
	l := Leader{}
	l.Index = index
	l.FlushMessages()
	l.m = map[*common.Vertex]*common.MessageEvent{}
	return l
}

func (leader *Leader) handleReceiveCommand(message []byte) common.MessageEvent {
	v := common.Vertex{leader.Index, id_count}
	id_count += 1
	newMessageEvent := common.MessageEvent{&v, message, []*common.Vertex{}}
	leader.m[&v] = &newMessageEvent
	return newMessageEvent
}

func (leader *Leader) AddToMessages(message *common.MessageEvent) {
	leader.messages = append(leader.messages, message)
}

func (leader *Leader) FlushMessages() {
	leader.messages = nil
}

func (leader *Leader) GetMessagesLen() int {
	return len(leader.messages)
}

func processMessageFromClient(m *nats.Msg, nc *nats.Conn, ctx context.Context, leader *Leader) {
	log.WithFields(log.Fields{
		"leaderId": leader.Index,
	}).Info("[BPAXOS] received message from client to leader")

	newMessage := leader.handleReceiveCommand(m.Data)
	sentMessage, err := json.Marshal(&newMessage)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("[BPAXOS] json marshal error while processing request to the leader node")
		return
	}
	subj := fmt.Sprintf("%s%d", common.LEADER_TO_PROPOSER, proposer_id)
	messenger.PublishNatsMessage(ctx, nc, subj, sentMessage)
	timer = time.NewTimer(time.Duration(common.PROPOSER_TIMEOUT_MILLISECONDS) * time.Millisecond)
	go leader.timeout(newMessage.VertexId, nc, ctx)

	proposer_id = (proposer_id + 1) % common.NUM_PROPOSERS
}

func (leader *Leader) timeout(v *common.Vertex, nc *nats.Conn, ctx context.Context) {
	<-timer.C
	log.Info("[BPAXOS] proposer timed out")
	mux.Lock()
	defer mux.Unlock()

	subj := fmt.Sprintf("%s%d", common.LEADER_TO_PROPOSER, proposer_id)
	sentMessage := leader.m[v]
	messenger.PublishNatsMessage(ctx, nc, subj, sentMessage)
	proposer_id = (proposer_id + 1) % common.NUM_PROPOSERS
	timer = time.NewTimer(time.Duration(common.PROPOSER_TIMEOUT_MILLISECONDS) * time.Millisecond)
	go leader.timeout(v, nc, ctx)

	log.Info("[BPAXOS] release timeout lock for leader")
}

func (leader *Leader) HandleConsensusDone(message *common.MessageEvent) {
	mux.Lock()
	defer mux.Unlock()
	timer.stop()
	delete(m, message.VertexId)
}

func StartLeader(ctx context.Context, nc *nats.Conn, leaderindex int) {
	l := NewLeader(leaderindex) // Hard Coded User Id.

	go func(nc *nats.Conn, leader *Leader) {
		natsMessage := make(chan *nats.Msg)
		err := messenger.SubscribeToInbox(ctx, nc, common.NATS_CONSENSUS_INITIATE_MSG, natsMessage, false)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe NATS_CONSENSUS_INITIATE_MSG")
		}

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-natsMessage:
				processMessageFromClient(natsMsg, nc, ctx, leader)
			}
		}
	}(nc, &l)

	go func(nc *nats.Conn, leader *Leader) {
		natsMessage := make(chan *nats.Msg)
		err := messenger.SubscribeToInbox(ctx, nc, common.PROPOSER_TO_REPLICA, natsMessage, false)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe NATS_CONSENSUS_INITIATE_MSG")
		}

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-natsMessage:
				data := common.MessageEvent{}
				json.Unmarshal(m.Data, &data)
				leader.HandleConsensusDone(&data)
			}
		}
	}(nc, &l)

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
