package leadernode

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

	// "strconv"

	log "github.com/sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

var (
	id_count    = 0
	proposer_id = 0 // This ID will determine which proposer to send to
	// timer        *time.Timer
	mux sync.Mutex
)

type Leader struct {
	Index       int
	messages    []*common.MessageEvent
	m           map[int]*common.MessageEvent
	t_map       map[int]*time.Timer
	q_map       map[int](chan bool)
	numberProps int
}

func (leader *Leader) SetNumProps() {
	// i, err := strconv.Atoi(os.Getenv("NUM_PROP"))
	// if err != nil {
	// 	log.WithFields(log.Fields{
	// 		"error": err.Error(),
	// 	}).Error("Failed to get Environment Variable")
	// } else {
	// 	leader.numberProps = i
	// }

	leader.numberProps = common.NUM_PROPOSERS
}

func NewLeader(index int) Leader {
	l := Leader{}
	l.Index = index
	log.SetOutput(os.Stdout)
	l.FlushMessages()
	l.m = make(map[int]*common.MessageEvent)
	l.t_map = make(map[int]*time.Timer) // timer map
	l.q_map = make(map[int](chan bool)) // quit map
	l.SetNumProps()
	return l
}

func (leader *Leader) handleReceiveCommand(message []byte) common.MessageEvent {
	v := common.Vertex{leader.Index, id_count}
	id_count += 1
	newMessageEvent := common.MessageEvent{&v, message, []*common.Vertex{}}

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
	leader.PrepareNewMessage(&newMessage)
	go leader.timeout(newMessage.VertexId, nc, ctx)

	proposer_id = (proposer_id + 1) % leader.numberProps
}

func (leader *Leader) checkMessageId(id int) bool {
	// Checks whether Message with ID equal to *id* is within leader.m
	// If so, return true
	// else return false
	// Only messages within leader.m are eligible to be proposed
	if _, ok := leader.m[id]; ok {
		return true
	}
	return false
}

func (leader *Leader) timeout(v *common.Vertex, nc *nats.Conn, ctx context.Context) {
	exit := false
	for {
		log.WithFields(log.Fields{"message id": v.Id}).Info("Started timeout routine in leader")
		select {
		case <-leader.t_map[v.Id].C:
			log.Info("[BPAXOS] leader waiting on proposer timed out")
			mux.Lock()

			subj := fmt.Sprintf("%s%d", common.LEADER_TO_PROPOSER, proposer_id)
			sentMessage, err := json.Marshal(leader.m[v.Id])
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("[BPAXOS] json marshal error while reproposing values from leader to proposers")
				return
			}
			messenger.PublishNatsMessage(ctx, nc, subj, sentMessage)
			proposer_id = (proposer_id + 1) % leader.numberProps
			leader.t_map[v.Id].Reset(time.Duration(common.PROPOSER_TIMEOUT_MILLISECONDS) * time.Millisecond)
			log.Info("[BPAXOS] release timeout lock for leader")
			mux.Unlock()
		case <-leader.q_map[v.Id]:
			exit = true
		}

		if exit {
			break
		}
	}
}

func (leader *Leader) HandleConsensusDone(message *common.MessageEvent) {
	// Consensus is reached. No need to repropose
	log.WithFields(log.Fields{
		"VertexId:": message.VertexId,
	}).Info("Leader is aware consensus Reached")
	leader.t_map[message.VertexId.Id].Stop()  // Stop the timer
	leader.q_map[message.VertexId.Id] <- true // Signal the timeout routine to quit
	delete(leader.t_map, message.VertexId.Id)
	delete(leader.m, message.VertexId.Id)
	delete(leader.q_map, message.VertexId.Id)
}

func (leader *Leader) PrepareNewMessage(message *common.MessageEvent) {
	timer := time.NewTimer(time.Duration(common.PROPOSER_TIMEOUT_MILLISECONDS) * time.Millisecond)
	// When the leader sends a new message, it needs to add the corresponding entries into its maps
	leader.t_map[message.VertexId.Id] = timer
	leader.q_map[message.VertexId.Id] = make(chan bool)
	leader.m[message.VertexId.Id] = message
}

func StartLeader(ctx context.Context, nc *nats.Conn, leaderindex int) {
	l := NewLeader(leaderindex) // Hard Coded User Id.

	go func(nc *nats.Conn, leader *Leader) {
		natsMessage := make(chan *nats.Msg, 100)
		err := messenger.SubscribeToInbox(ctx, nc, common.NATS_CONSENSUS_INITIATE_MSG, natsMessage, false)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe NATS_CONSENSUS_INITIATE_MSG")
		}
		debug.WriteToFile("L")

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
		debug.WriteToFile("L")

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-natsMessage:
				mux.Lock()
				data := common.MessageEvent{}
				json.Unmarshal(natsMsg.Data, &data)
				// Ensure that this data was originally proposed by this leader
				if data.VertexId.Index == leader.Index && leader.checkMessageId(data.VertexId.Id) {
					leader.HandleConsensusDone(&data)
					sentMessage, err := json.Marshal(&data)
					// Respond back to the client
					if err == nil {
						//fmt.Println("leader can publish a message to deps")
						messenger.PublishNatsMessage(ctx, nc, common.NATS_CONSENSUS_DONE_MSG, sentMessage)

					} else {
						fmt.Println("json marshal failed")
						fmt.Println(err.Error())
					}
				}
				mux.Unlock()
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
