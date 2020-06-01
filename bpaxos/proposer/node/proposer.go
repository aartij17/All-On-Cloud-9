package proposer

import (
	"All-On-Cloud-9/common"
	"All-On-Cloud-9/messenger"
	"All-On-Cloud-9/bpaxos/debug"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

var (
	requestQ     = make([]common.MessageEvent, 0)
	QueueTrigger = make(chan bool, common.F) // Max of F requests
	QueueRelease = make(chan bool)
	mux          sync.Mutex
	timer        *time.Timer
	timeout_quit = make(chan bool)
)

type Proposer struct {
	VoteCount  int
	Message    common.MessageEvent
	ProposerId int
}

func newProposer() Proposer {
	proposer := Proposer{}
	proposer.VoteCount = 0
	i, err := strconv.Atoi(os.Getenv("PROP_ID"))
	if err != nil {

		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to get Environment Variable")
	} else {
		proposer.ProposerId = i
	}
	return proposer
}

func (proposer *Proposer) HandleReceive(message *common.MessageEvent) {
	proposer.SendResult(message)
}

func (proposer *Proposer) SendResult(message *common.MessageEvent) {
	fmt.Println("send consensus result")
}

func (proposer *Proposer) processMessageFromLeader(data common.MessageEvent, nc *nats.Conn, ctx context.Context) {

	log.WithFields(log.Fields{
		"proposerId": proposer.ProposerId,
	}).Info("received message from leader to proposer")

	proposer.VoteCount = 0
	proposer.Message = data
	vertexId := data.VertexId
	// TODO: hardcoded the ip address for now, change that later
	consensusMessage := common.ConsensusMessage{VertexId: vertexId, ProposerId: proposer.ProposerId, Release: 0}
	sentMessage, err := json.Marshal(&consensusMessage)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("error marshal consensus message")
		return
	}
	messenger.PublishNatsMessage(ctx, nc, common.PROPOSER_TO_CONSENSUS, sentMessage)
	timer = time.NewTimer(time.Duration(common.CONSENSUS_TIMEOUT_MILLISECONDS) * time.Millisecond)
	go proposer.timeout()
}

func (proposer *Proposer) timeout() {
	select {
	case <-timer.C:
		log.Info("[BPAXOS] taking timeout lock for proposer")
		mux.Lock()
		
		if (len(requestQ) > 0) && (requestQ[0].VertexId.Id == proposer.Message.VertexId.Id) &&
			(requestQ[0].VertexId.Index == proposer.Message.VertexId.Index) {
			// Set Message Vertex to -1 so it will ignore any subsequent message related to this vertex
			proposer.Message.VertexId.Id = -1
			proposer.Message.VertexId.Index = -1
			QueueRelease <- true
			log.Error("Proposer timeout")
		}
		log.Info("[BPAXOS] release timeout lock for proposer")
                mux.Unlock()
	case <-timeout_quit:
		log.Info("[BPAXOS] Timeout not needed")
	}
}

func (proposer *Proposer) ProcessMessageFromConsensus(m *nats.Msg, nc *nats.Conn, ctx context.Context) {
	fmt.Println("Received consensus to proposer")
	data := common.Vertex{}
	err := json.Unmarshal(m.Data, &data)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err.Error(),
		}).Error("error unmarshal message from leader")
		return
	}

	if (data.Index == proposer.Message.VertexId.Index) && (data.Id == proposer.Message.VertexId.Id) {
		proposer.VoteCount += 1
		log.WithFields(log.Fields{
			"proposer vote count": proposer.VoteCount,
		}).Info("incrementing proposer vote count")
	}

	if proposer.VoteCount > common.F {
		timer.Stop()
		timeout_quit <- true
		replicaMessage, err := json.Marshal(&proposer.Message)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("error marshal proposer message")
			return
		}
		messenger.PublishNatsMessage(ctx, nc, common.PROPOSER_TO_REPLICA, replicaMessage)
		consensusMessage := common.ConsensusMessage{VertexId: proposer.Message.VertexId, ProposerId: proposer.ProposerId, Release: 1}
		sentConsensusMessage, err := json.Marshal(&consensusMessage)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err.Error(),
			}).Error("error marshal proposer release message")
			return
		}
		messenger.PublishNatsMessage(ctx, nc, common.PROPOSER_TO_CONSENSUS, sentConsensusMessage)

		proposer.Message.VertexId.Id = -1
		proposer.Message.VertexId.Index = -1
		QueueRelease <- true
	}
}

func StartProposer(ctx context.Context, nc *nats.Conn) {
	p := newProposer()

	go func(nc *nats.Conn, proposer *Proposer) {
		natsMessage := make(chan *nats.Msg, 100)
		subj := fmt.Sprintf("%s%d", common.LEADER_TO_PROPOSER, proposer.ProposerId)
		err := messenger.SubscribeToInbox(ctx, nc, subj, natsMessage, false)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe LEADER_TO_PROPOSER")
		}
		debug.WriteToFile("P")

		var (
			m *nats.Msg
		)
		for {
			select {
			case m = <-natsMessage:
				data := common.MessageEvent{}
				err := json.Unmarshal(m.Data, &data)
				if err != nil {
					log.WithFields(log.Fields{
						"err": err.Error(),
					}).Error("error unmarshal message from leader")
					return
				}
				requestQ = append(requestQ, data)
				QueueTrigger <- true
			}
		}
	}(nc, &p)

	go func(nc *nats.Conn, proposer *Proposer) {
		for {
			<-QueueTrigger
			proposer.processMessageFromLeader(requestQ[0], nc, ctx)
			<-QueueRelease
			requestQ = requestQ[1:]
		}
	}(nc, &p)

	go func(nc *nats.Conn, proposer *Proposer) {
		NatsMessage := make(chan *nats.Msg)
		sub := fmt.Sprintf("%s%d", common.CONSENSUS_TO_PROPOSER, proposer.ProposerId)
		err := messenger.SubscribeToInbox(ctx, nc, sub, NatsMessage, false)

		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("error subscribe CONSENSUS_TO_PROPOSER")
		}
		debug.WriteToFile("P")

		var (
			natsMsg *nats.Msg
		)
		for {
			select {
			case natsMsg = <-NatsMessage:
				mux.Lock()
				proposer.ProcessMessageFromConsensus(natsMsg, nc, ctx)
				mux.Unlock()
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
