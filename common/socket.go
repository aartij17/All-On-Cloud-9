package common

import (
	"os"
	"os/signal"
	log "github.com/Sirupsen/logrus"
	"github.com/nats-io/nats.go"
)

const (
	LeaderToDeps = "LeaderToDeps"
	DepsToLeader = "DepsToLeader"
	LeaderToProposer = "LeaderToProposer"
	ProposerToConsensus = "ProposerToConsensus"
	ConsensusToProposer = "ConsensusToProposer"
	ProposerToReplica = "ProposerToReplica"
)

type Socket struct {
	conn *nats.Conn
}

func (s *Socket) Connect(address string) bool {
	var err error

	natsOptions := nats.Options{
		Servers:        []string{address},
		AllowReconnect: true,
		MaxReconnect:   3,
		ReconnectWait:  10,
		//// TODO: these callbacks can be set in order to allow nats to perform failure handling.
		//ClosedCB:          nil,
		//DisconnectedErrCB: nil,
	}

	s.conn, err = natsOptions.Connect()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error connecting to the NATS server")
		return false
	}
	return true
}

func (s *Socket) Subscribe(subject string, response nats.MsgHandler ) {
	_, err := s.conn.Subscribe(subject, response)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error subscribing to the subject")
	}
}

func (s *Socket) Publish(subject string, request []byte ) {
	err := s.conn.Publish(subject,request)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("error publishing the nats request")
	}
}

func HandleInterrupt() {
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
