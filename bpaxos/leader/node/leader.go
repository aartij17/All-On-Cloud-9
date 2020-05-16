package leadernode

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

var (
	id_count    = 0
	proposer_id = 0 // This ID will determine which proposer to send to
)

type Leader struct {
	Index    int
	messages []*common.MessageEvent
}

func NewLeader(index int) Leader {
	l := Leader{}
	l.Index = index
	l.FlushMessages()
	return l
}

// func Union(a, b []*Vertex, m map[*Vertex]bool) []*Vertex {
// func (leader *Leader) Union() common.MessageEvent {

// 	newMessage := common.MessageEvent{}

// 	if len(leader.messages) > 0 {

// 		m := map[*common.Vertex]bool{}

// 		deps := leader.messages[0].Deps

// 		for _, item := range deps {
// 			m[item] = true
// 		}

// 		for _, mes := range leader.messages[1:] {
// 			for _, item := range mes.Deps {
// 				if _, ok := m[item]; !ok {
// 					deps = append(deps, item)
// 					m[item] = true
// 				}
// 			}
// 		}

// 		newMessage.VertexId = leader.messages[0].VertexId
// 		newMessage.Message = leader.messages[0].Message
// 		newMessage.Deps = deps
// 	}

// 	return newMessage
// }

func (leader *Leader) handleReceiveCommand(message []byte) common.MessageEvent {
	v := common.Vertex{leader.Index, id_count}
	id_count += 1
	newMessageEvent := common.MessageEvent{&v, message, []*common.Vertex{}}
	return newMessageEvent
}

// func (leader *Leader) HandleReceiveDeps() common.MessageEvent {

// 	newMessageEvent := leader.Union()
// 	return newMessageEvent

// }

func (leader *Leader) AddToMessages(message *common.MessageEvent) {
	leader.messages = append(leader.messages, message)
}

func (leader *Leader) FlushMessages() {
	leader.messages = nil
}

func (leader *Leader) GetMessagesLen() int {
	return len(leader.messages)
}

// func ProcessMessageFromDeps(m *nats.Msg, nc *nats.Conn, ctx context.Context, l Leader) {
// 	fmt.Println("Received deps to leader")
// 	data := common.MessageEvent{}
// 	json.Unmarshal(m.Data, &data)
// 	l.AddToMessages(&data)
// 	if l.GetMessagesLen() > common.F {
// 		newMessageEvent := l.HandleReceiveDeps()

// 		sentMessage, err := json.Marshal(&newMessageEvent)
// 		if err == nil {
// 			fmt.Println("leader can publish a message to proposer")
// 			messenger.PublishNatsMessage(ctx, nc, common.LEADER_TO_PROPOSER, sentMessage)
// 			// messenger.PublishNatsMessage(ctx, nc, common.NATS_CONSENSUS_DONE, sentMessage)

// 		} else {
// 			fmt.Println("json marshal failed")
// 			fmt.Println(err.Error())
// 		}
// 		// should we flush when it fails?
// 		l.FlushMessages()
// 	}
// }

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

	// messenger.PublishNatsMessage(ctx, nc, common.NATS_CONSENSUS_DONE, sentMessage)
	proposer_id = (proposer_id + 1) % common.NUM_PROPOSERS
}

func StartLeader(ctx context.Context, nc *nats.Conn, leaderindex int) {
	l := NewLeader(leaderindex) // Hard Coded User Id.

	go func(nc *nats.Conn, leader *Leader) {
		natsMessage := make(chan *nats.Msg)
		err := messenger.SubscribeToInbox(ctx, nc, common.NATS_CONSENSUS_INITIATE_MSG, natsMessage)

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
