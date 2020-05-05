package leadernode

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"encoding/json"
	"fmt"
	"context"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	log "github.com/Sirupsen/logrus"
)

var (
	id_count = 0
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
func (leader *Leader) Union() common.MessageEvent {

	newMessage := common.MessageEvent{}

	if len(leader.messages) > 0 {

		m := map[*common.Vertex]bool{}

		deps := leader.messages[0].Deps

		for _, item := range deps {
			m[item] = true
		}

		for _, mes := range leader.messages[1:] {
			for _, item := range mes.Deps {
				if _, ok := m[item]; !ok {
					deps = append(deps, item)
					m[item] = true
				}
			}
		}

		newMessage.VertexId = leader.messages[0].VertexId
		newMessage.Message = leader.messages[0].Message
		newMessage.Deps = deps
	}

	return newMessage
}

func (leader *Leader) HandleReceiveCommand(message string) common.MessageEvent {
	v := common.Vertex{leader.Index, id_count}
	id_count += 1
	newMessageEvent := common.MessageEvent{&v, message, []*common.Vertex{}}
	// fmt.Println(newMessageEvent.Message)  // STUB
	return newMessageEvent
	// send message to dependency
}

func (leader *Leader) HandleReceiveDeps() common.MessageEvent {

	newMessageEvent := leader.Union()
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

func StartLeader(ctx context.Context, leaderindex int) {
	l := NewLeader(leaderindex) // Hard Coded User Id.
	nc, err := messenger.NatsConnect(ctx)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error Leader connecting to nats server")
		return
	}

	go func(nc *nats.Conn, leader *Leader) {
		_, err = nc.Subscribe(common.DepsToLeader, func(m *nats.Msg) {
			fmt.Println("Received deps to leader")
			data := common.MessageEvent{}
			json.Unmarshal(m.Data, &data)
			l.AddToMessages(&data)
			if l.GetMessagesLen() > common.F {
				newMessageEvent := l.HandleReceiveDeps()
				
				sentMessage, err := json.Marshal(&newMessageEvent)
				if err == nil {
					fmt.Println("leader can publish a message to proposer")
					err = nc.Publish(common.LeaderToProposer, sentMessage)
					if err != nil {
						log.WithFields(log.Fields{
							"error": err.Error(),
						}).Error("error publish LeaderToProposer")
					}
				} else {
					fmt.Println("json marshal failed")
					fmt.Println(err.Error())
				}
				// should we flush when it fails?
				l.FlushMessages()
			}
		})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe DepsToLeader")
		}
	}(nc, &l)

	go func(nc *nats.Conn, leader *Leader) {
		_, err = nc.Subscribe(common.NATS_CONSENSUS_INITIATE_MSG, func(m *nats.Msg) {
			fmt.Println("Received client to leader")
			newMessage := leader.HandleReceiveCommand(string(m.Data))
			sentMessage, err := json.Marshal(&newMessage)
			if err == nil {
				fmt.Println("leader can publish a message to deps")
				err = nc.Publish(common.LeaderToDeps, sentMessage)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("error publish LeaderToDeps")
				}
			} else {
				fmt.Println("json marshal failed")
				fmt.Println(err.Error())
			}
		})
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe NATS_CONSENSUS_INITIATE_MSG")
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
